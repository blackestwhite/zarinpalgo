package zarinpalgo

import (
	"context"
	"testing"
	"github.com/google/uuid"
)

func TestSandboxPayment(t *testing.T) {
	// Generate a random merchant ID for testing
	merchantID := uuid.New().String()

	// Create a new Zarinpal client in sandbox mode
	zp := NewWithMode(merchantID, true)

	// Test sandbox URLs
	expectedBaseURL := "https://sandbox.zarinpal.com"
	if zp.APIBaseURL != expectedBaseURL+"/pg/v4/payment/" {
		t.Errorf("Expected API base URL %s, got %s", expectedBaseURL+"/pg/v4/payment/", zp.APIBaseURL)
	}
	if zp.PaymentBaseURL != expectedBaseURL+"/pg/StartPay/" {
		t.Errorf("Expected payment base URL %s, got %s", expectedBaseURL+"/pg/StartPay/", zp.PaymentBaseURL)
	}

	// Test payment creation
	amount := 10000 // 10,000 IRR
	description := "Test payment"
	callbackURL := "http://localhost:8080/callback"
	metadata := &Metadata{
		Email: "test@example.com",
		Mobile: "09123456789",
		OrderID: "TEST-ORDER-1",
	}

	payment, err := zp.NewPayment(context.Background(), amount, description, metadata, callbackURL, nil)
	if err != nil {
		t.Fatalf("Failed to create payment: %v", err)
	}

	// Verify payment creation response
	if payment.Authority == "" {
		t.Error("Expected non-empty authority token")
	}

	// Test payment URL generation
	paymentURL := zp.GetPaymentURL(payment.Authority)
	expectedURL := expectedBaseURL + "/pg/StartPay/" + payment.Authority
	if paymentURL != expectedURL {
		t.Errorf("Expected payment URL %s, got %s", expectedURL, paymentURL)
	}

	// Test payment verification
	// Note: In sandbox mode, we can't test actual payment verification as it requires user interaction
	// However, we can test the verification request structure
	status, err := zp.CheckPaymentStatus(context.Background(), amount, payment.Authority)
	if err != nil {
		// In sandbox mode, verification might fail as expected
		t.Logf("Payment verification failed as expected in sandbox mode: %v", err)
	} else {
		// If verification succeeds, check the response structure
		if status.Message == "" {
			t.Error("Expected non-empty status message")
		}
	}
}

func TestSandboxPaymentWithInvalidAmount(t *testing.T) {
	merchantID := uuid.New().String()
	zp := NewWithMode(merchantID, true)

	// Test payment with amount less than minimum (1000 Rials)
	amount := 999
	description := "Test payment with invalid amount"
	callbackURL := "http://localhost:8080/callback"

	_, err := zp.NewPayment(context.Background(), amount, description, nil, callbackURL, nil)
	if err == nil {
		t.Error("Expected error for amount less than 1000 Rials, got nil")
	}
}

func TestSandboxPaymentWithWages(t *testing.T) {
	merchantID := uuid.New().String()
	zp := NewWithMode(merchantID, true)

	// Test payment with wages
	amount := 20000
	description := "Test payment with wages"
	callbackURL := "http://localhost:8080/callback"
	wages := []Wage{
		{
			Iban:        "IR123456789012345678901234",
			Amount:      5000,
			Description: "Test wage payment",
		},
	}

	payment, err := zp.NewPayment(context.Background(), amount, description, nil, callbackURL, wages)
	if err != nil {
		t.Fatalf("Failed to create payment with wages: %v", err)
	}

	if payment.Authority == "" {
		t.Error("Expected non-empty authority token for payment with wages")
	}
}