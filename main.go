package main

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/shouxian92/SSDC-practical-checker/logger"
)

const (
	pollInterval = 30 * time.Minute
)

func main() {
	logger.NewInstance()
	logger.Info("application started")

	c := getAuthCookies()
	if len(c) < 2 {
		panic("main.go: don't have enough cookies to continue")
	}
	logger.Info("auth cookies obtained successfully: %v", c)

	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse(domain)
	jar.SetCookies(u, c)

	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// this can happen when our auth cookie expires, get new set of auth cookies
			c = getAuthCookies()
			jar.SetCookies(u, c)
			return http.ErrUseLastResponse
		},
	}

	formToken := ""
	ctx := &scriptBookingContext{
		Client:    client,
		XSRFToken: formToken,
	}

	for tick := range time.Tick(pollInterval) {
		hours, _, _ := tick.Clock()
		log.Printf("current hours is: %v\n", hours)
		// if hours > 15 || hours < 23 {
		// 	continue
		// }

		if len(formToken) <= 0 {
			formToken = initiateBookingFlow(client)
		}
		ctx.StartDate = time.Now()
		ctx.XSRFToken = formToken
		formToken = getAvailableTimeslots(*ctx)
	}

	logger.Info("application exiting")
}
