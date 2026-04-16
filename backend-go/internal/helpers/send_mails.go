package helpers

import (
	"fmt"
	"log"
)

type UserInfo struct {
	Email    string
	Username string
	ID       int32
}

// SendEmail is a global helper to send emails. 
// In a real implementation, this would use an SMTP client or an external service like SendGrid/Mailgun.
// For now, it logs the action.
func SendEmail(userInfo *UserInfo, template string, mailType string) error {
	recipient := "unknown"
	if userInfo != nil {
		recipient = fmt.Sprintf("%s <%s>", userInfo.Username, userInfo.Email)
	}

	mType := "general"
	if mailType != "" {
		mType = mailType
	}

	// Logic for sending email would go here
	log.Printf("[Email Service] Sending %s email to %s using template: %s", mType, recipient, template)
	
	// Example of actual implementation structure:
	// 1. Load SMTP config from environment
	// 2. Parse template with userInfo data
	// 3. Construct message
	// 4. Send using net/smtp or a client library

	return nil
}

// SendVerificationEmail is a convenience helper for OTP verification
func SendVerificationEmail(email, username, otp string) error {
	info := &UserInfo{Email: email, Username: username}
	template := fmt.Sprintf("Your verification code is: %s", otp)
	return SendEmail(info, template, "verification")
}
