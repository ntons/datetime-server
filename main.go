package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/ini.v1"
	"gopkg.in/natefinch/lumberjack.v2"
)

var cfg = struct {
	Main struct {
		Addr           string
		Certificate    string
		CertificateKey string
		LogPath        string
	}
	Log struct {
		MaxDataLen int
	}
}{}

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
	// parse configure
	if flag.Parse(); len(flag.Arg(0)) == 0 {
		log.Fatalf("Usage: %s path_to_configure", os.Args[0])
	}
	if err := ini.MapTo(&cfg, flag.Arg(0)); err != nil {
		log.Fatalf("parse fail: %v", err)
	}

	log.Printf("start %v", cfg)

	if cfg.Main.LogPath != "" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   cfg.Main.LogPath,
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     7,
		})
	}

	// listener
	l, err := net.Listen("tcp", cfg.Main.Addr)
	if err != nil {
		return
	}
	l = &Listener{Listener: l}

	// http
	http.Handle("/datetime", NewDateTimeHandler())
	http.Handle("/log", NewLogHandler())

	var s = &http.Server{}
	if len(cfg.Main.Certificate) == 0 || len(cfg.Main.CertificateKey) == 0 {
		go s.Serve(l)
	} else {
		go s.ServeTLS(l, cfg.Main.Certificate, cfg.Main.CertificateKey)
	}
	defer func() {
		if err = s.Shutdown(context.Background()); err != nil {
			log.Printf("shutdown fail: %v", err)
		}
	}()

	// waiting for term signal
	sig := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGPIPE)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
