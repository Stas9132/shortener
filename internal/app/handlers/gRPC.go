package handlers

import (
	"context"
	"errors"
	"github.com/Stas9132/shortener/internal/app/handlers/middleware"
	"github.com/Stas9132/shortener/internal/app/model"
	"github.com/Stas9132/shortener/internal/app/proto"
	"github.com/Stas9132/shortener/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
)

// GRPCAPI - ...
type GRPCAPI struct {
	logger.Logger
	proto.UnimplementedApiServer
	grpc.ServiceInfo
	m ModelAPI
}

// NewGRPCAPI - ...
func NewGRPCAPI(l logger.Logger, m ModelAPI) *GRPCAPI {
	return &GRPCAPI{Logger: l, m: m}
}

// Get - ...
func (a *GRPCAPI) Get(ctx context.Context, in *proto.ShortUrl) (*proto.OriginalURL, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

// Post - ...
func (a *GRPCAPI) Post(ctx context.Context, in *proto.OriginalURL) (*proto.ShortUrl, error) {
	u, err := url.Parse(in.GetURL())
	if err != nil {
		a.WithFields(map[string]interface{}{
			"error": err,
		}).Warn("url.Parse error")
	}
	response, err := a.m.Post(model.Request{URL: u}, middleware.GetIssuer(ctx).ID)

	if err != nil {
		if !errors.Is(err, model.ErrExist) {
			a.WithFields(map[string]interface{}{
				"error": err,
			}).Warn("model.Post error")
			return nil, err
		}
		return &proto.ShortUrl{URL: response.Result}, status.Error(codes.AlreadyExists, err.Error())
	}
	return &proto.ShortUrl{URL: response.Result}, nil
}

// PostBatch - ...
func (a *GRPCAPI) PostBatch(ctx context.Context, in *proto.Batch) (*proto.Batch, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostBatch not implemented")
}

// GetUserURLs - ...
func (a *GRPCAPI) GetUserURLs(ctx context.Context, in *proto.Empty) (*proto.Batch, error) {
	lu, err := a.m.GetUserURLs(ctx)
	if err != nil {
		if errors.Is(err, model.ErrUnauthorized) {
			a.WithFields(map[string]interface{}{
				"error": err,
			}).Warn("model.GetUserURLs error")
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	if len(lu) == 0 {
		return nil, status.Error(codes.NotFound, "")
	}

	return &proto.Batch{Records: func() []*proto.URLRecord {
		var res []*proto.URLRecord
		for _, t := range lu {
			res = append(res, &proto.URLRecord{
				OriginalURL: t.OriginalURL,
				ShortUrl:    t.ShortURL,
			})
		}
		return res
	}()}, nil
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
