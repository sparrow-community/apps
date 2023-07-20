package main

import (
	"fmt"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/registry/cache"
	"go-micro.dev/v4/selector"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func ReverseProxy() *http.ServeMux {
	rc := cache.New(registry.DefaultRegistry)

	m := http.NewServeMux()
	m.Handle("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		parts := strings.Split(request.URL.Path[1:], "/")
		services, err := rc.GetService(parts[0])
		if err != nil {
			err := errors.InternalServerError(ReverseProxyErr, "get upstream service [%s] error, %s", parts[0], err)
			logger.Error(err)
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		s, err := selector.Random(services)()
		if err != nil {
			err := errors.InternalServerError(ReverseProxyErr, "choice upstream service [%s] error %s", parts[0], err)
			logger.Error(err)
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		rp, err := url.Parse(fmt.Sprintf("http://%s", s.Address))
		if err != nil {
			err := errors.InternalServerError(ReverseProxyErr, "upstream service [%s] address error %s []", parts[0], err)
			logger.Error(err)
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		httputil.NewSingleHostReverseProxy(rp).ServeHTTP(writer, request)
	}))
	return m
}
