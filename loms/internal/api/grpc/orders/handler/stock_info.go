package api

import (
	"context"
	"errors"
	"route256/loms/internal/business/tool/converter"
	"route256/loms/internal/domain"
	desc "route256/loms/internal/pb/loms/v1"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (hdl *Implementation) StocksInfo(
	ctx context.Context, req *desc.StocksInfoRequest) (
	*desc.StocksInfoResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.StocksInfo")
	defer span.Finish()

	count, err := hdl.stockService.StocksInfo(ctx, domain.Sku(req.GetSku()))
	if err != nil {
		if errors.Is(err, domain.ErrStockNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, domain.ErrInternalServerError.Error())
	}

	convCount, err := converter.SafeInt64ToUint32(count)
	if err != nil {
		return nil, status.Error(codes.Internal, domain.ErrInternalServerError.Error())
	}

	return &desc.StocksInfoResponse{
		Count: convCount,
	}, nil
}
