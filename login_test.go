package main

import (
	"fmt"
	"io/ioutil"
	"log"
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

func TestReadCredentialsFromFile(t *testing.T) {
	originalTemplateContents, _ := ioutil.ReadFile(credentialsTemplateFile)

	// if original credentials file exists, back it up as well
	_, err := os.Stat(credentialsFile)
	hasCredentialsFile := !os.IsNotExist(err)
	originalCredContents := []byte{}
	if hasCredentialsFile {
		originalCredContents, _ = ioutil.ReadFile(credentialsFile)
		err = os.Remove(credentialsFile)
		if err != nil {
			log.Fatalf("setup failed to remove credentials file: %v", err)
		}
	}

	t.Run("RenamesTemplateToActual", func(t *testing.T) {
		getCredentialsFromFile()

		_, err := os.Stat(credentialsTemplateFile)
		assert.True(t, os.IsNotExist(err))

		_, err = os.Stat(credentialsFile)
		assert.Nil(t, err)
	})

	t.Run("SuccessfullyRead", func(t *testing.T) {
		err := ioutil.WriteFile(credentialsFile, []byte(fmt.Sprintf(validFileContents, username, password)), 0644)

		if err != nil {
			log.Fatalf("test failed to write to file: %v", err)
		}

		c := getCredentialsFromFile()

		assert.Equal(t, c.Username, username)
		assert.Equal(t, c.Password, password)
	})

	t.Run("EmptyFileReturnsEmptyCredentials", func(t *testing.T) {
		err := ioutil.WriteFile(credentialsFile, []byte(""), 0644)
		if err != nil {
			log.Fatalf("test failed to write to file: %v", err)
		}

		c := getCredentialsFromFile()

		assert.Empty(t, c)
	})

	if hasCredentialsFile {
		err := ioutil.WriteFile(credentialsFile, originalCredContents, 0644)
		if err != nil {
			log.Fatalf("test cleanup failed to write back to file: %v", err)
		}
	} else {
		err := os.Remove(credentialsFile)
		if err != nil {
			log.Fatalf("cleanup failed to remove credentials file: %v", err)
		}
	}

	err = ioutil.WriteFile(credentialsTemplateFile, originalTemplateContents, 0644)
	if err != nil {
		log.Fatalf("test cleanup failed to write back to file: %v", err)
	}
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
