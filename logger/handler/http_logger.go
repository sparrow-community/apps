package handler

import "net/http"

type HttpHandler struct {
}

var DefaultHttpHandler = &HttpHandler{}

func (l *HttpHandler) Welcome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Welcome"))
}
