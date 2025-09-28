package provider

import (
	lomsclient "route256/cart/internal/adapter/client/loms"
	desc "route256/cart/internal/pb/loms/v1"
	"time"

	"google.golang.org/grpc"
)

func ProvideLOMSClient(conn *grpc.ClientConn, timeout time.Duration) (*lomsclient.Client, error) {
	orderClient := desc.NewOrdersClient(conn)
	stockClient := desc.NewStocksClient(conn)

	return lomsclient.New(orderClient, stockClient, timeout), nil
}
