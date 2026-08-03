package grpcclient

import (
	"context"
	"net"
	"testing"

	paymentv1 "github.com/dipeshmalviya/subscription-billing-platform/gen/go/payment/v1"
	"google.golang.org/grpc"
)

type testPaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
}

func (s *testPaymentServer) ChargeCustomer(ctx context.Context, req *paymentv1.ChargeCustomerRequest) (*paymentv1.ChargeCustomerResponse, error) {
	return &paymentv1.ChargeCustomerResponse{
		PaymentId:   "payment-123",
		Status:      paymentv1.PaymentStatus_PAYMENT_STATUS_SUCCEEDED,
		ProviderRef: "provider-abc",
	}, nil
}

func TestPaymentClient_ChargeCustomer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(server, &testPaymentServer{})
	go server.Serve(listener)
	defer server.Stop()

	client, err := NewPaymentClient(listener.Addr().String())
	if err != nil {
		t.Fatalf("NewPaymentClient failed: %v", err)
	}
	defer client.Close()

	resp, err := client.ChargeCustomer(context.Background(), "inv-1", "cust-1", 1000, "USD", "idem-1")
	if err != nil {
		t.Fatalf("ChargeCustomer failed: %v", err)
	}
	if resp.PaymentId != "payment-123" {
		t.Fatalf("expected payment id payment-123, got %q", resp.PaymentId)
	}
	if resp.Status != paymentv1.PaymentStatus_PAYMENT_STATUS_SUCCEEDED {
		t.Fatalf("expected status PAYMENT_STATUS_SUCCEEDED, got %v", resp.Status)
	}
	if resp.ProviderRef != "provider-abc" {
		t.Fatalf("expected provider ref provider-abc, got %q", resp.ProviderRef)
	}
}
