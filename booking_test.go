package main

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

type testSessionContext struct {
	SessionID     string
	SessionNumber string
	Date          string
}

func TestInitiateBookingFlow(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, domain+"/User/Booking/AddBooking?bookingType=PL", func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(http.StatusOK, `<html><input name="`+xsrfTokenName+`" value="foo"/></html>`)
		resp.Header = http.Header{}
		resp.Header.Set("Set-Cookie", xsrfTokenName+"=baz;")
		return resp, nil
	})

	mockClient := &http.Client{}
	res := initiateBookingFlow(mockClient)
	assert.Equal(t, "foo", res)
}

func TestGetAvailableTimeslots(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	testCases := []struct {
		name               string
		availableTimeslots []testSessionContext // to generate anchor tags for testing
	}{
		{"HasOneTimeslot", []testSessionContext{{SessionID: "0", SessionNumber: "1", Date: "5/10/2019 10:00:00 PM"}}},
	}

	resp := &http.Response{}
	httpmock.RegisterResponder(http.MethodPost, domain+"/User/Booking/AddBooking", func(req *http.Request) (*http.Response, error) {
		return resp, nil
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp = getCustomResponse(tc.availableTimeslots)
			ctx := &scriptBookingContext{
				StartDate: time.Now(),
				Client:    &http.Client{},
				XSRFToken: "",
			}
			resultingXSRF := getAvailableTimeslots(*ctx)
			assert.Equal(t, "anotherXSRF", resultingXSRF)
		})
	}
}

func TestBuildAvailableTimeslots_Success(t *testing.T) {
	testCases := []struct {
		name               string
		availableTimeslots []testSessionContext
		expectedDates      []string
	}{
		{
			name: "HasOneTimeslot",
			availableTimeslots: []testSessionContext{
				{SessionID: "0", SessionNumber: "1", Date: "5/10/2019 10:00:00 PM"},
			},
			expectedDates: []string{"05 Oct 2019 (Sat)"},
		},
		{
			name:               "HasNoTimeslots",
			availableTimeslots: []testSessionContext{},
			expectedDates:      []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := getCustomResponse(tc.availableTimeslots)
			res := buildAvailableTimeslots(resp)
			assert.Equal(t, len(tc.availableTimeslots), len(res))

			for i, r := range res {
				assert.Equal(t, tc.availableTimeslots[i].SessionID, r.SessionID)
				assert.Equal(t, sessionNumberToStartTime[tc.availableTimeslots[i].SessionNumber], r.StartTime)
				assert.Equal(t, tc.expectedDates[i], r.Date)
			}
		})
	}
}

func getCustomResponse(ctx []testSessionContext) *http.Response {
	anchorTags := ""
	for _, c := range ctx {
		anchorTags = anchorTags + `<a href="#" class="slotBooking" id="` + c.SessionID + `_` + c.SessionNumber + `_` + c.Date + `"></a>`
	}
	log.Println(`<html>` + anchorTags + `<input class="` + xsrfTokenName + `" value="anotherXSRF"/></html>`)
	return httpmock.NewStringResponse(http.StatusOK, `<html>`+anchorTags+`<input name="`+xsrfTokenName+`" value="anotherXSRF"/></html>`)
}
