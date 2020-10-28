package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const (
	xsrfTokenName     = "__RequestVerificationToken"
	sessionCookieName = ".AspNet.ApplicationCookie"
	sessionIdName     = "ASP.NET_SessionId"
	domain            = "https://www.ssdcl.com.sg"
)

func getXSRFTokens(r http.Response) (string, string) {
	cookieXSRFToken := ""
	for _, cookie := range r.Cookies() {
		if cookie.Name == xsrfTokenName {
			cookieXSRFToken = cookie.Value
			break
		}
	}

	formXSRFToken := ""
	tokenizer := html.NewTokenizer(r.Body)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.SelfClosingTagToken {
			token := tokenizer.Token()
			if token.Data == "input" {
				isXSRFToken := false
				for _, attr := range token.Attr {
					if attr.Key == "name" && attr.Val == xsrfTokenName {
						isXSRFToken = true
					}
				}

				if isXSRFToken {
					for _, attr := range token.Attr {
						if attr.Key == "value" {
							formXSRFToken = attr.Val
						}
					}
				}
			}
		}
	}
	r.Body.Close()

	if len(strings.TrimSpace(cookieXSRFToken)) <= 0 && len(strings.TrimSpace(formXSRFToken)) <= 0 {
		panic("XSRF token could not be obtained")
	}

	return cookieXSRFToken, formXSRFToken
}

func getCredentialsFromFile() credentials {
	b, err := ioutil.ReadFile(".credentials.json")

	if err != nil {
		fmt.Printf("unable to read credentials file: %v\n", err)
	}

	c := credentials{}
	err = json.Unmarshal(b, &c)

	if err != nil {
		fmt.Printf("failed to parse file contents: %v\n", err)
	}

	return c
}

// return the http auth cookie that can be used to make website calls
func login() (*http.Cookie, error) {
	creds := getCredentialsFromFile()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, _ := http.NewRequest(http.MethodGet, domain+"/User/Login", nil)
	r, _ := client.Do(req)
	cookieXSRFToken, formXSRFToken := getXSRFTokens(*r)

	postData := url.Values{}
	postData.Add("Password", creds.Password)
	postData.Add("UserName", creds.Username)
	postData.Add(xsrfTokenName, formXSRFToken)

	req, _ = http.NewRequest(http.MethodPost, domain+"/Account/Login", strings.NewReader(postData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "__RequestVerificationToken="+cookieXSRFToken+";")

	r, _ = client.Do(req)
	defer r.Body.Close()

	if len(r.Cookies()) <= 0 {
		panic("no cookies in response")
	}

	for _, cookie := range r.Cookies() {
		if cookie.Name == ".AspNet.ApplicationCookie" {
			return cookie, nil
		}
	}

	return &http.Cookie{}, fmt.Errorf("no session cookie found")
}

func main() {

	//c, err := login()

	// if err == nil {
	// 	fmt.Printf("successfully obtained cookie: %v\n", c)
	// } else {
	// 	fmt.Printf("failed to get cookie: %v\n", err)
	// }

	jar, _ := cookiejar.New(nil)
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:   sessionCookieName,
		Value:  "wv0bmRzIlkV2iahURGdSx5AlAMArYSq3eRX2zLboSJH9fatbmUgex_bqpVzb-pc_0355IbXhx_z76Y7Hia206_RYQ1e1a0UeE-2Iz9JdWTP6S-pXI0_C6Vo3Q3eMgDM3FFS3_ueUa9y_JpSjBB3HkiwwFShK2s9yFRyD9K7uy46g_w6NdiW46TrC3TuXaswjsdZwCQ3A7x22NXwelLTMaE_VJDlF5cMKC8HPaooEMShMqvPvtEcsfQa0pwMpBhBY4bKVBHclzqoGUDHH2AjfNrte4tNXwnuQvn7hbGj6RhLBsI8tAR1wIDlPWqCp3q6-6-cK0GW85bTZHdGlWyrDdqLFgfHrPDsXyLtxUVVupkMq9B-LXhr8QE4q80xEwYwhDrVojMi-qKyjh4CQLvpSKA",
		Path:   "/",
		Domain: ".ssdcl.com.sg",
	}
	cookies = append(cookies, cookie)
	u, _ := url.Parse(domain)
	jar.SetCookies(u, cookies)
	client := &http.Client{
		Jar: jar,
	}

	/*c := time.Tick(1 * time.Minute)
	for _ = range c {
		fmt.Print("Polling for SSDC")
	}*/
}
