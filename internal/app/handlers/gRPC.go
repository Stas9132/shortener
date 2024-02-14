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
	s, err := a.m.GetRoot(in.GetURL())
	if err != nil {
		if !errors.Is(err, model.ErrNotFound) {
			a.WithFields(map[string]interface{}{
				"error": err,
			}).Warn("model.GetRoot")
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.NotFound, "")
	}
	return &proto.OriginalURL{URL: s}, nil
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
	i, err := a.m.PostBatch(func() model.Batch {
		var res model.Batch
		for _, record := range in.GetRecords() {
			res = append(res, struct {
				CorrelationID string `json:"correlation_id"`
				OriginalURL   string `json:"original_url,omitempty"`
				ShortURL      string `json:"short_url"`
			}{CorrelationID: "", OriginalURL: record.GetOriginalURL(), ShortURL: ""})
		}
		return res
	}())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	for j := 0; j <= i; j++ {
		in.GetRecords()[i].OriginalURL = ""
	}

	return in, nil
}

// GetUserURLs - ...
func (a *GRPCAPI) GetUserURLs(ctx context.Context, in *proto.Empty) (*proto.Batch, error) {
	lu, err := a.m.GetUserURLs(ctx)
	if err != nil {
		if !errors.Is(err, model.ErrUnauthorized) {
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
	_, err := a.m.DeleteUserUrls(func() model.BatchDelete {
		var res []string
		for _, record := range in.GetRecords() {
			res = append(res, record.ShortUrl)
		}
		return res
	}())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &proto.Empty{}, nil
}

// Ping - ...
func (a *GRPCAPI) Ping(ctx context.Context, in *proto.Empty) (*proto.Empty, error) {
	err := a.m.GetPing()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.Empty{}, nil
}

// GetStats - ...
func (a *GRPCAPI) GetStats(ctx context.Context, in *proto.Empty) (*proto.Stats, error) {
	stats, err := a.m.GetStats()
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &proto.Stats{
		URLs:  int32(stats.Urls),
		Users: int32(stats.Users),
	}, nil
}
