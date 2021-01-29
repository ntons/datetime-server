package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type LogHandler struct{}

func NewLogHandler() *LogHandler {
	return &LogHandler{}
}

func SeparateIpFromAddr(addr string) string {
	if i := strings.IndexByte(addr, ':'); i < 0 {
		return addr
	} else {
		return addr[:i]
	}
}

func (h LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if cfg.Log.MaxDataLen > 0 && len(b) > cfg.Log.MaxDataLen {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("log %s %s", SeparateIpFromAddr(r.RemoteAddr), b)
}
