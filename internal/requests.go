package internal

// SlackRequestBody is an object containing data for Slack notification endpoint.
type SlackRequestBody struct {
	Message string `validate:"required" json:"message"`
}

// SMSRequestBody is an object containing data for SMS notification endpoint.
type SMSRequestBody struct {
	Message      string `validate:"required" json:"message"`
	SendToNumber string `validate:"required,e164" json:"send_to_number"`
}

// MailRequestBody is an object containing data for mail notification endpoint.
type MailRequestBody struct {
	Message string `validate:"required" json:"message"`
	SendTo  string `validate:"required,email" json:"send_to"`
	Subject string `validate:"required" json:"subject"`
}
