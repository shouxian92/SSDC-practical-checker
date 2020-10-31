package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func bodyToBytes(r *http.Response) []byte {
	b, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	rc := ioutil.NopCloser(bytes.NewBuffer(b))
	r.Body = rc
	return b
}
