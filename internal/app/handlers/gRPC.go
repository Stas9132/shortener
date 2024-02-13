package handlers

import (
	"context"
	"github.com/Stas9132/shortener/internal/app/proto"
	"github.com/Stas9132/shortener/internal/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCAPI - ...
type GRPCAPI struct {
	logger.Logger
	proto.UnimplementedApiServer
}

// NewGRPCAPI - ...
func NewGRPCAPI(l logger.Logger) *GRPCAPI {
	return &GRPCAPI{Logger: l}
}

// Get - ...
func (a *GRPCAPI) Get(ctx context.Context, in *proto.ShortUrl) (*proto.OriginalURL, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

// Post - ...
func (a *GRPCAPI) Post(ctx context.Context, in *proto.OriginalURL) (*proto.ShortUrl, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Post not implemented")
}

// PostBatch - ...
func (a *GRPCAPI) PostBatch(ctx context.Context, in *proto.Batch) (*proto.Batch, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostBatch not implemented")
}

// GetUserURLs - ...
func (a *GRPCAPI) GetUserURLs(ctx context.Context, in *proto.Empty) (*proto.Batch, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserURLs not implemented")
}

// DeleteUserURLs - ...
func (a *GRPCAPI) DeleteUserURLs(ctx context.Context, in *proto.Batch) (*proto.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUserURLs not implemented")
}

// Ping - ...
func (a *GRPCAPI) Ping(ctx context.Context, in *proto.Empty) (*proto.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

// GetStats - ...
func (a *GRPCAPI) GetStats(ctx context.Context, in *proto.Empty) (*proto.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStats not implemented")
}
