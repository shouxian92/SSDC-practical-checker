package notification

import "github.com/shouxian92/SSDC-practical-checker/structures"

// EmailContext is used to pass information to generate the required payload for sending the email
type EmailContext struct {
	to string
	ts []structures.Timeslot
}

// SendEmail will trigger a notification in the form of an email
func SendEmail(ctx EmailContext) {

}
