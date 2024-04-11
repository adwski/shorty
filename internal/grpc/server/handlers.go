//nolint:wrapcheck // using gstatus.Error() to return grpc errors
package server

import (
	"context"
	"errors"

	g "github.com/adwski/shorty/internal/grpc"
	"github.com/adwski/shorty/internal/model"
	"github.com/adwski/shorty/internal/services/resolver"
	"github.com/adwski/shorty/internal/services/shortener"
	"github.com/adwski/shorty/internal/session"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

// ErrRequestCtx indicates error while getting info from request context.
var (
	ErrRequestCtx = "request context error"
)

// Stats returns storage statistics.
func (srv *Server) Stats(ctx context.Context, _ *g.StatsRequest) (*g.StatsResponse, error) {
	resp, err := srv.statusSvc.Stats(ctx)
	if err != nil {
		return nil, gstatus.Errorf(codes.Internal, "internal error occured")
	}
	return &g.StatsResponse{
		Urls:  int64(resp.URLs),
		Users: int64(resp.Users),
	}, nil
}

// Resolve retrieves original URL of corresponding shortened URL.
func (srv *Server) Resolve(ctx context.Context, r *g.ResolveRequest) (*g.ResolveResponse, error) {
	reqID, ok := session.GetRequestID(ctx)
	if !ok {
		srv.logger.Error("request id was not provided in context")
		return nil, gstatus.Errorf(codes.Internal, ErrRequestCtx)
	}

	result, err := srv.resolverSvc.Resolve(ctx, r.Path)
	srv.logger.With(
		zap.String("path", r.Path),
		zap.String("orig", result),
		zap.String("id", reqID),
		zap.Error(err),
	).Debug("resolve called")

	if err != nil {
		switch {
		case errors.Is(err, resolver.ErrInvalidPath):
			return nil, gstatus.Errorf(codes.InvalidArgument, "invalid path")
		case errors.Is(err, model.ErrNotFound):
			return nil, gstatus.Errorf(codes.NotFound, "path is not found")
		case errors.Is(err, model.ErrDeleted):
			return nil, gstatus.Errorf(codes.FailedPrecondition, "path is deleted")
		default:
			return nil, gstatus.Error(codes.Internal, "internal error occurred")
		}
	}
	return &g.ResolveResponse{OriginalUrl: result}, nil
}

// Shorten generates short URL for provided original URL and stores it.
// Short URL is returned back.
func (srv *Server) Shorten(ctx context.Context, r *g.ShortenRequest) (*g.ShortenResponse, error) {
	u, reqID, err := session.GetUserAndReqID(ctx)
	if err != nil {
		srv.logger.Error(ErrRequestCtx, zap.Error(err))
		return nil, gstatus.Errorf(codes.Internal, ErrRequestCtx)
	}

	result, err := srv.shortenerSvc.Shorten(ctx, u, r.OriginalUrl)
	srv.logger.With(
		zap.String("result", result),
		zap.String("orig", r.OriginalUrl),
		zap.String("id", reqID),
		zap.String("userID", u.ID),
		zap.Error(err),
	).Debug("shorten called")

	if err != nil {
		switch {
		case errors.Is(shortener.ErrInvalidURL, err),
			errors.Is(shortener.ErrUnsupportedURLScheme, err):
			return nil, gstatus.Error(codes.InvalidArgument, err.Error())

		case errors.Is(model.ErrConflict, err):
			err = gstatus.Error(codes.FailedPrecondition, "url conflict")
		default:
			return nil, gstatus.Error(codes.Internal, "internal error")
		}
	}
	return &g.ShortenResponse{ShortUrl: result}, err // nil or url conflict
}

// ShortenBatch shortens batch of original URLs. It returns batch of short URLs
// that can be matched with originals using correlation ID.
func (srv *Server) ShortenBatch(ctx context.Context, r *g.ShortenBatchRequest) (*g.ShortenBatchResponse, error) {
	u, reqID, err := session.GetUserAndReqID(ctx)
	if err != nil {
		srv.logger.Error(ErrRequestCtx, zap.Error(err))
		return nil, gstatus.Errorf(codes.Internal, ErrRequestCtx)
	}

	batchURLs := make([]shortener.BatchURL, 0, len(r.BatchUrl))
	for i := range r.BatchUrl {
		batchURLs = append(batchURLs, shortener.BatchURL{
			ID:  r.BatchUrl[i].CorrelationId,
			URL: r.BatchUrl[i].OriginalUrl,
		})
	}

	shortURLs, err := srv.shortenerSvc.ShortenBatch(ctx, u, batchURLs)
	srv.logger.With(
		zap.Int("shortURLs", len(shortURLs)),
		zap.String("id", reqID),
		zap.String("userID", u.ID),
		zap.Bool("newUser", u.IsNew()),
		zap.Error(err),
	).Debug("ShortenBatch called")
	if err != nil {
		return nil, gstatus.Error(codes.Internal, "internal error")
	}

	var resp g.ShortenBatchResponse
	resp.BatchUrl = make([]*g.ShortURL, 0, len(shortURLs))
	for i := range shortURLs {
		resp.BatchUrl = append(resp.BatchUrl, &g.ShortURL{
			CorrelationId: shortURLs[i].ID,
			ShortUrl:      shortURLs[i].Short,
		})
	}
	return &resp, nil
}

// DeleteBatch schedules list of URLs for deletion. Actual deletion is done asynchronously.
func (srv *Server) DeleteBatch(ctx context.Context, r *g.DeleteBatchRequest) (*g.DeleteBatchResponse, error) {
	u, reqID, err := session.GetUserAndReqID(ctx)
	if err != nil {
		srv.logger.Error(ErrRequestCtx, zap.Error(err))
		return nil, gstatus.Error(codes.Internal, ErrRequestCtx)
	}

	err = srv.shortenerSvc.DeleteBatch(ctx, u, r.Hashes)
	srv.logger.With(
		zap.String("id", reqID),
		zap.String("userID", u.ID),
		zap.Error(err),
	).Debug("DeleteBatch called")
	if err != nil {
		switch {
		case errors.Is(shortener.ErrUnauthorized, err):
			return nil, gstatus.Error(codes.Unauthenticated, "unauthorized")
		case errors.Is(shortener.ErrEmptyBatch, err):
			return nil, gstatus.Error(codes.InvalidArgument, "empty batch")
		default:
			return nil, gstatus.Error(codes.Internal, "delete error")
		}
	}
	return nil, gstatus.Error(codes.OK, "accepted")
}

// GetAll returns all URLs created by single user.
func (srv *Server) GetAll(ctx context.Context, _ *g.GetAllRequest) (*g.GetAllResponse, error) {
	u, reqID, err := session.GetUserAndReqID(ctx)
	if err != nil {
		srv.logger.Error(ErrRequestCtx, zap.Error(err))
		return nil, gstatus.Errorf(codes.Internal, ErrRequestCtx)
	}

	urls, err := srv.shortenerSvc.GetAll(ctx, u)
	srv.logger.With(
		zap.Int("urls", len(urls)),
		zap.String("id", reqID),
		zap.String("userID", u.ID),
		zap.Error(err),
	).Debug("getAll called")
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, gstatus.Errorf(codes.NotFound, "no urls")
		}
		return nil, gstatus.Error(codes.Internal, "internal error")
	}

	var resp g.GetAllResponse
	resp.Urls = make([]*g.URL, 0, len(urls))
	for i := range urls {
		resp.Urls = append(resp.Urls, &g.URL{
			ShortUrl:    urls[i].Short,
			OriginalUrl: urls[i].Orig,
		})
	}
	return &resp, nil
}
