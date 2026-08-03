package grpcclient

import (
	"context"
	"fmt"

	paymentv1 "github.com/dipeshmalviya/subscription-billing-platform/gen/go/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentClient struct {
	conn   *grpc.ClientConn
	client paymentv1.PaymentServiceClient
}

func NewPaymentClient(addr string) (*PaymentClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial payment service: %w", err)
	}

	return &PaymentClient{
		conn:   conn,
		client: paymentv1.NewPaymentServiceClient(conn),
	}, nil
}

func (c *PaymentClient) ChargeCustomer(
	ctx context.Context,
	invoiceID string,
	customerID string,
	amountCents int64,
	currency string,
	idempotencyKey string,
) (*paymentv1.ChargeCustomerResponse, error) {
	req := &paymentv1.ChargeCustomerRequest{
		InvoiceId:      invoiceID,
		CustomerId:     customerID,
		AmountCents:    amountCents,
		Currency:       currency,
		IdempotencyKey: idempotencyKey,
	}
	return c.client.ChargeCustomer(ctx, req)
}

func (c *PaymentClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
