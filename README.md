# ZarinpalGo 
Unofficial Zarinpal REST API integration for Go

## Installation
```bash
go get github.com/blackestwhite/zarinpalgo
```

## Usage

### Initialize the Client
```go
zp := zarinpalgo.New("YOUR-MERCHANT-ID")
```

### Create a New Payment
Use `NewPayment` to initiate a payment request:

```go
// Optional metadata
metadata := &zarinpalgo.Metadata{
    Email:   "customer@example.com",
    Mobile:  "09123456789",
    OrderID: "ORDER-123",
}

// Optional wage payments
wages := []zarinpalgo.Wage{
    {
        Iban:        "IR123456789",
        Amount:      1000,
        Description: "Service fee",
    },
}

// Create payment request
response, err := zp.NewPayment(
    context.Background(),
    1000000, // amount in Rials
    "Payment for order #123",
    metadata,        // can be nil
    "https://your-callback-url.com",
    wages,          // can be nil
)

if err != nil {
    log.Fatal(err)
}

// Get payment URL
paymentURL := zp.GetPaymentURL(response.Authority)
// Redirect user to paymentURL
```

### Verify Payment
After the user is redirected back to your callback URL, use `CheckPaymentStatus` to verify the payment:

```go
// Get authority and amount from the callback request
authority := "YOUR-AUTHORITY-TOKEN"
amount := 1000000 // same amount as payment request

// Check payment status
status, err := zp.CheckPaymentStatus(context.Background(), amount, authority)
if err != nil {
    log.Fatal(err)
}

if status.IsSuccessful {
    if status.IsRepeated {
        fmt.Println("Payment was successful but verified before")
    } else {
        fmt.Println("Payment was successful")
    }
    fmt.Printf("RefID: %d\n", status.RefID)
} else {
    fmt.Printf("Payment failed: %s\n", status.Message)
}
```

## Features
- Easy to use API client for Zarinpal payment gateway
- Support for payment metadata
- Support for wage payments
- User-friendly payment verification
- Proper error handling and status checking
- Context support for timeouts and cancellation

## Response Types

### PaymentStatus
The `CheckPaymentStatus` method returns a user-friendly `PaymentStatus` struct:
```go
type PaymentStatus struct {
    IsSuccessful bool   // true if payment was successful
    IsRepeated   bool   // true if payment was verified before
    RefID        int    // payment reference ID
    Message      string // status message
}
```

## Error Handling
The package provides proper error handling for API responses and network issues. Always check the returned error and status message for proper handling of edge cases.