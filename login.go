package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/shouxian92/SSDC-practical-checker/logger"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func getCredentialsFromFile() credentials {
	b, err := ioutil.ReadFile(".credentials.json")

	if err != nil {
		logger.LogError("unable to read credentials file: %v\n", err)
	}

	c := credentials{}
	err = json.Unmarshal(b, &c)

	if err != nil {
		logger.LogError("failed to parse file contents: %v\n", err)
	}

	return c
}

// return the http auth cookie and XSRF token cookie that can be used to make website calls
func getAuthCookies() []*http.Cookie {
	creds := getCredentialsFromFile()
	cookies := []*http.Cookie{}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, _ := http.NewRequest(http.MethodGet, domain+"/User/Login", nil)
	r, _ := client.Do(req)
	cookieXSRFToken, formXSRFToken := getTokens(r)
	cookies = append(cookies, cookieXSRFToken)

	postData := url.Values{}
	postData.Add("Password", creds.Password)
	postData.Add("UserName", creds.Username)
	postData.Add(xsrfTokenName, formXSRFToken)

	req, _ = http.NewRequest(http.MethodPost, domain+"/Account/Login", strings.NewReader(postData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookieXSRFToken)

	r, err := client.Do(req)

	if err != nil {
		defer r.Body.Close()
	}

	if len(r.Cookies()) <= 0 {
		panic("no cookies in response")
	}

	for _, cookie := range r.Cookies() {
		if cookie.Name == sessionCookieName {
			cookies = append(cookies, cookie)
		}
	}

	return cookies
}
