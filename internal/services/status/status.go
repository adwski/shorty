package status

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type Pingable interface {
	Ping(context.Context) error
}

type Service struct {
	store Pingable
	log   *zap.Logger
}

type Config struct {
	Storage any
}

func New(cfg *Config) (*Service, error) {
	if store, ok := cfg.Storage.(Pingable); ok {
		return &Service{store: store}, nil
	}
	return nil, fmt.Errorf("storage is not Pingable")
}

func (svc *Service) PingStorage(w http.ResponseWriter, req *http.Request) {
	err := svc.store.Ping(req.Context())
	if err != nil {
		svc.log.Error("storage ping unsuccessful", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
