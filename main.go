package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type credentials struct {
	username string
	password string
}

// return the http cookie that can be used to make website calls
func login() (*http.Cookie, error) {
	dat, err := ioutil.ReadFile("./.credentials")

	if err != nil {
		fmt.Printf("unable to read credentials file: %v", err)
	}

	creds := &credentials{}
	json.Unmarshal(dat, &creds)

	// Grab the RequestVerificationToken by loading the login page
	loginURL := "https://www.ssdcl.com.sg/User/Login"
	r, _ := http.Get(loginURL)
	d, _ := goquery.NewDocumentFromReader(r.Body)
	defer r.Body.Close()

	xsrfToken := ""
	//<input name="__RequestVerificationToken" type="hidden" value="Pl3gJ-33V0P-x_QrfjucRQcuZltUCAFlyYRaUoy0Zpx4n2_2cEOy8W-olYvG-IBnz8QnfRO8jsLJeT_zH-5jfCc91vXpQqkbkXqFZBqLsHo1" />
	d.Find("input").Each(func(index int, element *goquery.Selection) {
		if n, exists := element.Attr("name"); exists && n == "__RequestVerificationToken" {
			xsrfToken, _ = element.Attr("value")
		}
	})

	if len(strings.TrimSpace(xsrfToken)) <= 0 {
		panic("XSRF token could not be obtained")
	}

	// Make actual login
	formData := url.Values{
		"__RequestVerificationToken": {xsrfToken},
		"UserName":                   {creds.username},
		"Password":                   {creds.password},
	}

	r, _ = http.PostForm("https://www.ssdcl.com.sg/Account/Login", formData)
	for _, cookie := range r.Cookies() {
		if cookie.Name == ".AspNet.ApplicationCookie" {
			return cookie, nil
		}
	}

	return &http.Cookie{}, fmt.Errorf("no session cookie found")
}

func main() {

	/*c := time.Tick(1 * time.Minute)
	for _ = range c {
		fmt.Print("Polling for SSDC")
	}*/
}
