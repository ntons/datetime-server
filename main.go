package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/ini.v1"
)

type IniHTTP struct {
	Addr string
	Path string
}

type IniSSL struct {
	Enabled        bool
	Certificate    string
	CertificateKey string
}

type IniDatetime struct {
	Format string
}

var cfg = struct {
	HTTP     *IniHTTP
	SSL      *IniSSL
	Datetime *IniDatetime
}{}

func NamedFormat(s string) string {
	switch s {
	case "ANSIC":
		return time.ANSIC
	case "UnixDate":
		return time.UnixDate
	case "RubyDate":
		return time.RubyDate
	case "RFC822":
		return time.RFC822
	case "RFC822Z":
		return time.RFC822Z
	case "RFC850":
		return time.RFC850
	case "RFC1123":
		return time.RFC1123
	case "RFC1123Z":
		return time.RFC1123Z
	case "RFC3339":
		return time.RFC3339
	case "RFC3339Nano":
		return time.RFC3339Nano
	case "Kitchen":
		return time.Kitchen
	case "Stamp":
		return time.Stamp
	case "StampMilli":
		return time.StampMilli
	case "StampMicro":
		return time.StampMicro
	case "StampNano":
		return time.StampNano
	default:
		return s
	}
}

func HandleDatetime(w http.ResponseWriter, r *http.Request) {
	switch cfg.Datetime.Format {
	case "Unix":
		w.Write([]byte(fmt.Sprintf("%d", time.Now().Unix())))
	case "UnixNano":
		w.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	default:
		w.Write([]byte(time.Now().Format(NamedFormat(cfg.Datetime.Format))))
	}
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
	var err error

	flag.Parse()
	if len(flag.Arg(0)) == 0 {
		log.Fatalf("Usage: %s path_to_configure", os.Args[0])
	}

	// parse configure
	if err = ini.MapTo(&cfg, flag.Arg(0)); err != nil {
		log.Fatalf("parse fail: %v", err)
	}

	// listen
	var l net.Listener
	if l, err = net.Listen("tcp", cfg.HTTP.Addr); err != nil {
		log.Fatalf("listen fail: %v", err)
	}
	l = &Listener{Listener: l}

	// http
	http.HandleFunc(cfg.HTTP.Path, HandleDatetime)
	var s = &http.Server{}
	if !cfg.SSL.Enabled {
		go s.Serve(l)
	} else {
		go s.ServeTLS(l, cfg.SSL.Certificate, cfg.SSL.CertificateKey)
	}
	sig := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGPIPE)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	if err = s.Shutdown(context.Background()); err != nil {
		log.Printf("http server shutdown: %v", err)
	}
}
