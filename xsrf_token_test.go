package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Given a list of cookies, i must expect that the request verification token is returned
func TestGetXSRFCookie_Success(t *testing.T) {
	cookies := []*http.Cookie{}
	xsrfCookie := &http.Cookie{
		Name:  xsrfTokenName,
		Value: "foo",
	}
	cookies = append(cookies, xsrfCookie)
	res := getXSRFCookie(cookies)

	assert.Equal(t, xsrfTokenName, res.Name)
	assert.Equal(t, "foo", res.Value)
}

// Given a http response, i must obtain an entropy string which represents the hidden XSRF input
func TestGetXSRFForm_Success(t *testing.T) {
	const mockResponseBody = "<html><input type=\"hidden\" name=\"__RequestVerificationToken\" value=\"foobar\" /></html>"
	mockResponse := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(mockResponseBody)),
	}

	res := getXSRFForm(mockResponse)
	assert.Greater(t, len(res), 0)
	assert.Equal(t, "foobar", res)
}

func TestGetXSRFForm_NoInputWithXSRFTokenName(t *testing.T) {
	const mockResponseBody = "<html><input type=\"hidden\" value=\"foobar\" /></html>"
	mockResponse := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(mockResponseBody)),
	}

	res := getXSRFForm(mockResponse)
	assert.Equal(t, len(res), 0)
	assert.Equal(t, "", res)
}
