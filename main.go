package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"gopkg.in/ini.v1"
)

type IniHTTP struct {
	Addr string
	Path string

	Certificate    string
	CertificateKey string
}

var cfg = struct {
	HTTP *IniHTTP
}{}

// REQ: t0=xxx
// RES: t0=xxx&t1=xxx&t2=xxx
func Handle(w http.ResponseWriter, r *http.Request) {
	t1 := time.Now().UnixNano() / 1e3
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	t0 := r.Form.Get("t0")
	if len(t0) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	buf.WriteString(fmt.Sprintf("t0=%s&t1=%d&t2=", t0, t1))
	t2 := time.Now().UnixNano() / 1e3
	buf.WriteString(strconv.FormatInt(t2, 10))
	w.Write(buf.Bytes())
}

type Listener struct {
	net.Listener
}

func (l Listener) Accept() (conn net.Conn, err error) {
	if conn, err = l.Listener.Accept(); err != nil {
		return
	}
	if err = conn.(*net.TCPConn).SetNoDelay(true); err != nil {
		return
	}
	return
}

func main() {
	flag.Parse()
	if len(flag.Arg(0)) == 0 {
		log.Fatalf("Usage: %s path_to_configure", os.Args[0])
	}

	// parse configure
	if err := ini.MapTo(&cfg, flag.Arg(0)); err != nil {
		log.Fatalf("parse fail: %v", err)
	}

	// listener
	l, err := net.Listen("tcp", cfg.HTTP.Addr)
	if err != nil {
		return
	}
	l = &Listener{Listener: l}

	// http
	http.HandleFunc(cfg.HTTP.Path, Handle)
	var s = &http.Server{}
	if len(cfg.HTTP.Certificate) == 0 || len(cfg.HTTP.CertificateKey) == 0 {
		go s.Serve(l)
	} else {
		go s.ServeTLS(l, cfg.HTTP.Certificate, cfg.HTTP.CertificateKey)
	}
	defer func() {
		if err = s.Shutdown(context.Background()); err != nil {
			log.Printf("shutdown fail: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGPIPE)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
