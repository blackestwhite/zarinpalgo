package zarinpalgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Zarinpal struct {
	MerchantID     string
	APIBaseURL     string
	PaymentBaseURL string
	client         *http.Client
}

// PaymentStatus represents the result of a payment verification
type PaymentStatus struct {
	IsSuccessful bool
	IsRepeated   bool
	RefID        int
	Message      string
}

type PaymentRequest struct {
	MerchantID  string    `json:"merchant_id"`
	Amount      int       `json:"amount"`
	Description string    `json:"description"`
	Metadata    *Metadata `json:"metadata,omitempty"`
	CallbackURL string    `json:"callback_url"`
	Wages       []Wage    `json:"wages,omitempty"`
}

type PaymentVerificationRequest struct {
	MerchantID string `json:"merchant_id"`
	Amount     int    `json:"amount"`
	Authority  string `json:"authority"`
}

type Metadata struct {
	Email   string `json:"email"`
	Mobile  string `json:"mobile"`
	OrderID string `json:"order_id"`
}

type Wage struct {
	Iban        string `json:"iban"`
	Amount      int    `json:"amount"`
	Description string `json:"description"`
}

type PaymentCreationResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Authority string `json:"authority"`
	FeeType   string `json:"fee_type"`
	Fee       int    `json:"fee"`
}

type PaymentVerificationResponse struct {
	Code     int    `json:"code"` // 100 means payment was successful, 101 means the payment was successful and is verified before
	Message  string `json:"message"`
	CardHash string `json:"card_hash"`
	CardPan  string `json:"card_pan"`
	RefID    int    `json:"ref_id"`
	FeeType  string `json:"fee_type"`
	Fee      int    `json:"fee"`
}

type BaseResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors json.RawMessage `json:"errors"`
}

type ErrorResponse struct {
	Message     string        `json:"message"`
	Code        int           `json:"code"`
	Validations []interface{} `json:"validations"`
}

// PaymentResult constants
const (
	PaymentCodeSuccess         = 100 // Payment was successful
	PaymentCodeAlreadyVerified = 101 // Payment was successful and verified before
)

// New creates a new Zarinpal client with the given merchant ID
func New(merchantID string) *Zarinpal {
	return NewWithMode(merchantID, false)
}

// NewWithMode creates a new Zarinpal client with the given merchant ID and sandbox mode
func NewWithMode(merchantID string, sandbox bool) *Zarinpal {
	baseURL := "https://payment.zarinpal.com"
	if sandbox {
		baseURL = "https://sandbox.zarinpal.com"
	}

	return &Zarinpal{
		MerchantID:     merchantID,
		APIBaseURL:     baseURL + "/pg/v4/payment/",
		PaymentBaseURL: baseURL + "/pg/StartPay/",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewPayment initiates a new payment request
func (z *Zarinpal) NewPayment(ctx context.Context, amount int, description string, metadata *Metadata, callbackURL string, wages []Wage) (paymentCreationResponse PaymentCreationResponse, err error) {
	paymentRequestBody := PaymentRequest{
		MerchantID:  z.MerchantID,
		Amount:      amount,
		Description: description,
		Metadata:    metadata,
		CallbackURL: callbackURL,
		Wages:       wages,
	}

	marshalled, err := json.Marshal(paymentRequestBody)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", z.APIBaseURL+"request.json", bytes.NewBuffer(marshalled))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := z.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	rawMessage, err := checkResponse(bodyBytes)
	if err != nil {
		return
	}

	err = json.Unmarshal(rawMessage, &paymentCreationResponse)
	if err != nil {
		return
	}

	return
}

// VerifyPayment verifies a payment using authority and amount
func (z *Zarinpal) VerifyPayment(ctx context.Context, amount int, authority string) (paymentVerificationResponse PaymentVerificationResponse, err error) {
	paymentVerificationRequestBody := PaymentVerificationRequest{
		MerchantID: z.MerchantID,
		Amount:     amount,
		Authority:  authority,
	}

	marshalled, err := json.Marshal(paymentVerificationRequestBody)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", z.APIBaseURL+"verify.json", bytes.NewBuffer(marshalled))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := z.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	rawMessage, err := checkResponse(bodyBytes)
	if err != nil {
		return
	}

	err = json.Unmarshal(rawMessage, &paymentVerificationResponse)
	if err != nil {
		return
	}

	return
}

// CheckPaymentStatus verifies a payment and returns a user-friendly status
func (z *Zarinpal) CheckPaymentStatus(ctx context.Context, amount int, authority string) (PaymentStatus, error) {
	verification, err := z.VerifyPayment(ctx, amount, authority)
	if err != nil {
		return PaymentStatus{
			IsSuccessful: false,
			Message:      err.Error(),
		}, err
	}

	status := PaymentStatus{
		Message: verification.Message,
		RefID:   verification.RefID,
	}

	// Check if payment was successful
	switch verification.Code {
	case PaymentCodeSuccess:
		status.IsSuccessful = true
		status.IsRepeated = false
	case PaymentCodeAlreadyVerified:
		status.IsSuccessful = true
		status.IsRepeated = true
	default:
		status.IsSuccessful = false
		status.IsRepeated = false
	}

	return status, nil
}

// GetPaymentURL generates the payment URL from an authority token
func (z *Zarinpal) GetPaymentURL(authority string) string {
	return z.PaymentBaseURL + authority
}

func checkResponse(body []byte) (rawMessage json.RawMessage, err error) {
	var baseResponse BaseResponse
	err = json.Unmarshal(body, &baseResponse)
	if err != nil {
		return
	}

	isEmptyErrors := string(baseResponse.Errors) == "[]" || string(baseResponse.Errors) == "{}" || string(baseResponse.Errors) == ""

	if !isEmptyErrors {
		var errorResponse ErrorResponse
		err = json.Unmarshal(baseResponse.Errors, &errorResponse)
		if err != nil {
			return
		}
		err = fmt.Errorf("error code: %d, error: %s", errorResponse.Code, errorResponse.Message)
		return
	}

	return baseResponse.Data, nil
}
