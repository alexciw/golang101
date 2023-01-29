package _1_simplewebserver

import (
	"net/http"
	"testing"
	"time"
)

func TestHttpServer(t *testing.T) {
	s := http.Server{
		Addr:        ":8080",
		ReadTimeout: time.Second * 30,
	}

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}
