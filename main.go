package main

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var (
	pollInterval = 30 * time.Minute
)

func main() {
	go bindToHerokuPort()
	initLogger()

	zap.S().Info("application started")

	client := initHTTPClient()
	ctx := &scriptBookingContext{
		Client:    client,
		XSRFToken: "",
	}

	if im, ok := os.LookupEnv("INTERVAL_MINUTES"); ok {
		m, err := strconv.Atoi(im)
		if err == nil && m > 0 {
			zap.S().Infof("setting poll interval to %v minutes", m)
			pollInterval = time.Duration(m) * time.Minute
		}
	}

	for tick := range time.Tick(pollInterval) {
		zap.S().Info("starting clock interval")
		hours, _, _ := tick.UTC().Clock()
		if hours > 15 && hours < 23 {
			zap.S().Info("not going to run at this interval since it is sleepy time")
			continue
		}

		if len(ctx.XSRFToken) <= 0 {
			ctx.XSRFToken = initiateBookingFlow(client)
		}
		ctx.StartDate = time.Now()
		ctx.XSRFToken = getAvailableTimeslots(*ctx)
	}

	zap.S().Info("application exiting")
	err := zap.L().Sync()

	if err != nil {
		log.Fatalf("unable to flush zap logger: %v", err)
	}
}

func initLogger() {
	logger, err := zap.NewProduction()
	zap.ReplaceGlobals(logger)
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}
}

func initHTTPClient() *http.Client {
	zap.S().Info("initializing http client")
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse(domain)

	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			c := getAuthCookies()

			if len(c) < 2 {
				zap.S().Error("don't have enough cookies to continue")
				return nil
			}

			zap.S().Infof("auth cookies obtained successfully: %v", c)
			jar.SetCookies(u, c)
			return nil
		},
	}
}

func bindToHerokuPort() {
	port := os.Getenv("PORT")

	if len(port) == 0 {
		return
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
