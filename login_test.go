package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	username          = "foo"
	password          = "bar"
	validFileContents = "{\"username\": \"%s\", \"password\":\"%s\"}"
)

func TestMain(m *testing.M) {
	originalContents, _ := ioutil.ReadFile(credentialsTemplateFile)
	code := m.Run()
	os.Remove(credentialsFile)
	ioutil.WriteFile(credentialsTemplateFile, originalContents, 0644)
	os.Exit(code)
}

func TestReadCredentialsFile_RenamesTemplateToActual(t *testing.T) {
	getCredentialsFromFile()

	_, err := os.Stat(credentialsTemplateFile)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(credentialsFile)
	assert.Nil(t, err)
}

func TestReadCredentialsFile_Successful(t *testing.T) {
	ioutil.WriteFile(credentialsFile, []byte(fmt.Sprintf(validFileContents, username, password)), 0644)

	c := getCredentialsFromFile()

	assert.Equal(t, c.Username, username)
	assert.Equal(t, c.Password, password)
}

func TestReadCredentialsFile_InvalidFileReturnsEmptyCredentials(t *testing.T) {
	ioutil.WriteFile(credentialsFile, []byte(""), 0644)
	c := getCredentialsFromFile()

	assert.Empty(t, c)
}

func TestGetAuthCookies_Successful(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, domain+"/User/Login", func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(http.StatusOK, "<html><input name=\""+xsrfTokenName+"\" value=\"foo\"/></html>")
		resp.Header = http.Header{}
		resp.Header.Set("Set-Cookie", xsrfTokenName+"=baz;")
		return resp, nil
	})
	httpmock.RegisterResponder(http.MethodPost, domain+"/Account/Login", func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(http.StatusFound, "")
		resp.Header = http.Header{}
		resp.Header.Set("Set-Cookie", sessionCookieName+"=loggedin;")
		resp.Header.Set("Location", "some.random.place")
		return resp, nil
	})

	cookies := getAuthCookies()
	expectedCookies := map[string]string{
		xsrfTokenName:     "baz",
		sessionCookieName: "loggedin",
	}

	for _, c := range cookies {
		v, exists := expectedCookies[c.Name]
		assert.True(t, exists)
		assert.Equal(t, v, c.Value)
	}
}
