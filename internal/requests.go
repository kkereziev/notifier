package internal

// SlackRequestBody is an object used to validate request parameters for Slack notification endpoint.
type SlackRequestBody struct {
	Message string `validate:"required" json:"message"`
}

// SMSRequestBody is an object used to validate request parameters for SMS notification endpoint.
type SMSRequestBody struct {
	Message      string `validate:"required" json:"message"`
	SendToNumber string `validate:"required,e164" json:"send_to_number"`
}
