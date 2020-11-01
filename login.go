package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/shouxian92/SSDC-practical-checker/logger"
)

const (
	credentialsFile         = ".credentials.json"
	credentialsTemplateFile = ".credentials.template.json"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func getCredentialsFromFile() credentials {
	c := credentials{}

	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		// check template file instead
		renameErr := os.Rename(credentialsTemplateFile, credentialsFile)

		if renameErr != nil {
			logger.LogError("failed to rename template file: %v", err)
		}
	}

	b, err := ioutil.ReadFile(credentialsFile)

	if err != nil {
		logger.LogError("unable to read credentials file: %v\n", err)
		return c
	}

	err = json.Unmarshal(b, &c)

	if err != nil {
		logger.LogError("failed to parse file contents: %v\n", err)
	}

	if len(c.Username) == 0 || len(c.Password) == 0 {
		logger.LogError("username or password is empty in credentials file.")
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

	postData := url.Values{
		"Password":    {creds.Password},
		"UserName":    {creds.Username},
		xsrfTokenName: {formXSRFToken},
	}

	req, _ = http.NewRequest(http.MethodPost, domain+"/Account/Login", strings.NewReader(postData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookieXSRFToken)

	r, err := client.Do(req)

	if err == nil {
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
