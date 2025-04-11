package grpc

import (
	"context"
	"order-pick-up-point/internal/models/entity"
	"order-pick-up-point/internal/storage/db"
	"order-pick-up-point/pkg/logger"
)

type PvzService interface {
	GetPVZList(ctx context.Context) ([]entity.Pvz, error)
}

type pvzServiceImp struct {
	repo   db.PvzRepository
	logger logger.Logger
}

func NewPvzService(repo db.PvzRepository, log logger.Logger) PvzService {
	return &pvzServiceImp{
		repo:   repo,
		logger: log,
	}
}

func (s *pvzServiceImp) GetPVZList(ctx context.Context) ([]entity.Pvz, error) {
	pvzs, err := s.repo.GetListOfPvzs(ctx)
	if err != nil {
		s.logger.Errorw("get PVZ list",
			"error", err,
		)
		return nil, err
	}
	return pvzs, nil
}
