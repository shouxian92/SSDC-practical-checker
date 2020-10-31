package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

// grabs xsrf cookie from the response headers
func getXSRFCookie(requestCookies []*http.Cookie) *http.Cookie {
	c := &http.Cookie{}
	for _, cookie := range requestCookies {
		if cookie.Name == xsrfTokenName {
			c = cookie
			break
		}
	}

	if len(c.Value) <= 0 {
		fmt.Println("unable to obtain cookie XSRF token")
	}

	return c
}

// crawl a byte buffer that is assumed to be HTML to obtain the hidden XSRF token input
func getXSRFForm(r *http.Response) string {
	formXSRFToken := ""
	rc := ioutil.NopCloser(bytes.NewBuffer(bodyToBytes(r)))
	tokenizer := html.NewTokenizer(rc)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.SelfClosingTagToken {
			token := tokenizer.Token()
			if token.Data == "input" {
				attributes := make(map[string]string)
				for _, attr := range token.Attr {
					attributes[attr.Key] = attr.Val
				}
				className, exists := attributes["name"]
				if exists && className == xsrfTokenName {
					formXSRFToken = attributes["value"]
				}
			}
		}
	}

	if len(formXSRFToken) <= 0 {
		log.Println("unable to obtain XSRF token from hidden input")
	}

	return formXSRFToken
}

// return tokens needed for making requests
func getTokens(r *http.Response) (*http.Cookie, string) {
	cookieToken := getXSRFCookie(r.Cookies())
	formToken := getXSRFForm(r)

	if len(cookieToken.Value) <= 0 && len(strings.TrimSpace(formToken)) <= 0 {
		panic("XSRF tokens could not be obtained")
	}

	return cookieToken, formToken
}
