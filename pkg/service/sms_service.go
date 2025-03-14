package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"log/slog"
)

// SmsSender defines the behavior of an SMS sending service.
type SmsSender interface {
	SendSms(recipient string, message string) (string, error)
}

// TwilioSmsSender implements SmsSender using Twilio's REST API.
type TwilioSmsSender struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	HTTPClient *http.Client
	Logger     *slog.Logger
}

// NewTwilioSmsSender creates a new TwilioSmsSender with the provided configuration.
func NewTwilioSmsSender(accountSID string, authToken string, fromNumber string, logger *slog.Logger) *TwilioSmsSender {
	return &TwilioSmsSender{
		AccountSID: accountSID,
		AuthToken:  authToken,
		FromNumber: fromNumber,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Logger:     logger,
	}
}

// SendSms sends an SMS using Twilio's API.
func (senderInstance *TwilioSmsSender) SendSms(recipient string, message string) (string, error) {
	formData := url.Values{}
	formData.Set("To", recipient)
	formData.Set("From", senderInstance.FromNumber)
	formData.Set("Body", message)

	apiEndpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", senderInstance.AccountSID)
	requestInstance, requestError := http.NewRequest(http.MethodPost, apiEndpoint, strings.NewReader(formData.Encode()))
	if requestError != nil {
		senderInstance.Logger.Error("Failed to create Twilio request", "error", requestError)
		return "", requestError
	}
	requestInstance.SetBasicAuth(senderInstance.AccountSID, senderInstance.AuthToken)
	requestInstance.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	responseInstance, responseError := senderInstance.HTTPClient.Do(requestInstance)
	if responseError != nil {
		senderInstance.Logger.Error("Twilio request error", "error", responseError)
		return "", responseError
	}
	defer responseInstance.Body.Close()

	responseBody, _ := io.ReadAll(responseInstance.Body)
	if responseInstance.StatusCode >= 300 {
		senderInstance.Logger.Error("Twilio API returned error", "status", responseInstance.StatusCode, "body", string(responseBody))
		return "", fmt.Errorf("twilio API error: %s", string(responseBody))
	}

	return string(responseBody), nil
}
