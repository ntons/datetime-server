package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type DateTimeHandler struct{}

func NewDateTimeHandler() *DateTimeHandler {
	return &DateTimeHandler{}
}

func (h DateTimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
