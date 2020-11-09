package notifications

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/shouxian92/SSDC-practical-checker/logger"
	"github.com/shouxian92/SSDC-practical-checker/structures"
)

// EmailContext is used to pass information to generate the required payload for sending the email
type EmailContext struct {
	To        string
	Timeslots []structures.Timeslot
}

type sendRequest struct {
	Type     string      `json:"type"`
	To       string      `json:"to"`
	Message  string      `json:"message"`
	Metadata interface{} `json:"metadata"`
}

// SendEmail will trigger a notification in the form of an email
func SendEmail(ctx EmailContext) {
	if len(ctx.Timeslots) <= 0 {
		return
	}

	b, _ := json.Marshal(&sendRequest{
		Type:     "driving",
		To:       ctx.To,
		Message:  "",
		Metadata: ctx.Timeslots,
	})

	resp, err := http.Post(os.Getenv("EMAIL_SERVER")+"/send", "application/json", bytes.NewBuffer(b))

	if err != nil {
		logger.Error("there was an error making the email request: %v", err)
		return
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		logger.Info("email sent successfully")
		break
	case http.StatusBadRequest:
		logger.Info("bad request: %v", resp.Body)
		break
	default:
		logger.Info("unexpected error happened: (%v) %v", resp.StatusCode, resp.Body)
		break
	}

	defer resp.Body.Close()
}
