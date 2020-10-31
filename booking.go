package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shouxian92/SSDC-practical-checker/logger"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

// this map represents the session number to the start time of the session
var (
	sessionNumberToStartTime = map[string]string{
		"1": "08:15AM",
		"2": "10:30AM",
		"3": "01:05PM",
		"4": "03:20PM",
		"5": "06:10PM",
		"6": "08:20PM",
	}
)

// a struct to hold data for representing the available timeslots
type timeslot struct {
	SessionID string
	StartTime string
	Date      string
}

// context that belongs to the website
type ssdcBookingContext struct {
	XSRFToken             string
	SlotID                int
	SelectedSessionNumber int
	SellBundleID          string
	IsOrientation         bool
	BookingType           string
	SelectedDate          time.Time
	SelectedSessionType   string
	SelectedLocation      string
	CarModelID            int
	IsFiRequired          bool
}

type scriptBookingContext struct {
	StartDate time.Time
	Client    *http.Client
	XSRFToken string
	Logger    *zap.SugaredLogger
}

func buildAvailableTimeslots(r *http.Response) []timeslot {
	rc := ioutil.NopCloser(bytes.NewBuffer(bodyToBytes(r)))
	tokenizer := html.NewTokenizer(rc)
	availableTimeslots := []timeslot{}

	timeslotSet := make(map[string]bool)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "a" {
				attributes := make(map[string]string)
				for _, attr := range token.Attr {
					attributes[attr.Key] = attr.Val
				}
				className, exists := attributes["class"]
				if exists && className == "slotBooking" {
					timeslotSet[attributes["id"]] = true
				}
			}
		}
	}

	for id := range timeslotSet {
		sessionContext := strings.Split(id, "_")

		if len(sessionContext) > 5 {
			logger.LogWarn("not a valid timeslot, the session ID did not contain 3 items: %v", id)
		}

		// the date format of the element is not according to any RFC format
		// we need to truncate AM/PM away from the string for parsing
		dateStr := sessionContext[2][0 : len(sessionContext[2])-3]
		date, err := time.Parse("2/01/2006 15:04:05", dateStr)

		if err != nil {
			logger.LogWarn("failed to parse timestamp. timestamp was %v. error: %v\n", sessionContext[2], err)
			continue
		}

		ts := &timeslot{
			SessionID: sessionContext[0],
			StartTime: sessionNumberToStartTime[sessionContext[1]],
			Date:      date.Format("02 Jan 2006 (Mon)"),
		}

		availableTimeslots = append(availableTimeslots, *ts)
	}

	return availableTimeslots
}

// makes a POST call to AddBooking with an XSRF token
// returns a new XSRF token after the booking has been checked
func getAvailableTimeslots(ctx scriptBookingContext) string {
	context := &ssdcBookingContext{
		XSRFToken:             ctx.XSRFToken,
		SlotID:                0,
		SelectedSessionNumber: 0,
		SelectedDate:          ctx.StartDate,
	}
	formData := buildFormData(*context)

	req, _ := http.NewRequest(http.MethodPost, domain+"/User/Booking/AddBooking", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := ctx.Client.Do(req)

	if err != nil || resp.StatusCode == http.StatusFound {
		logger.LogError("error making request (resp code: %v): %v", resp.StatusCode, err)
		return ""
	}

	defer resp.Body.Close()
	newXSRFToken := getXSRFForm(resp)
	availableTimeslots := buildAvailableTimeslots(resp)

	logger.LogInfo("available timeslots: %v", availableTimeslots)

	return newXSRFToken
}

// makes a GET call to return the very first XSRF token
func initiateBookingFlow(client *http.Client) string {
	req, _ := http.NewRequest(http.MethodGet, domain+"/User/Booking/AddBooking?bookingType=PL", nil)
	resp, err := client.Do(req)
	if err != nil {
		logger.LogError("error navigating to booking page: %v", err)
	}
	formToken := getXSRFForm(resp)
	resp.Body.Close()
	logger.LogInfo("GET /AddBooking completed successfully. XSRF token: %v\n", formToken)
	return formToken
}

func buildFormData(ctx ssdcBookingContext) url.Values {
	d := &url.Values{}
	d.Add(xsrfTokenName, ctx.XSRFToken)
	d.Add("SlotId", strconv.Itoa(ctx.SlotID))
	d.Add("SelectedSessionNumber", strconv.Itoa(ctx.SelectedSessionNumber))
	d.Add("SelectedDate", ctx.SelectedDate.Format("02 Jan 2006"))

	// allow some level of context checking later
	//d.Add("SellBundleId", ctx.SellBundleID)
	//d.Add("IsOrientation", strings.Title(strconv.FormatBool(ctx.IsOrientation)))
	//d.Add("BookingType", ctx.BookingType)
	//d.Add("SelectedSessionType", ctx.SelectedSessionType)
	//d.Add("SelectedLocation", ctx.SelectedLocation)
	//d.Add("CarModelId", strconv.Itoa(ctx.CarModelID))
	//d.Add("IsFiRequired", strings.Title(strconv.FormatBool(ctx.IsFiRequired)))
	//d.Add("checkEligibility", ctx.checkEligibility)

	// defaults
	d.Add("SellBundleId", "00000000-0000-0000-0000-000000000000")
	d.Add("IsOrientation", "False")
	d.Add("BookingType", "PL")
	d.Add("SelectedSessionType", "N")
	d.Add("SelectedLocation", "Woodlands")
	d.Add("CarModelId", "1")
	d.Add("IsFiRequired", "False")
	d.Add("checkEligibility", "CHECK_SLOT_AVAILABLE")

	return *d
}
