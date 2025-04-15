package http

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"order-pick-up-point/internal/errs"
	"order-pick-up-point/internal/metrics"
	"order-pick-up-point/internal/models/entity"
	"order-pick-up-point/internal/storage/db"
	"order-pick-up-point/pkg/logger"
	"strings"
	"time"
)

type PvzService interface {
	CreatePvz(ctx context.Context, city string) (string, error)
	GetPvzsInfo(ctx context.Context, page, limit int, startDate, endDate *time.Time) ([]entity.PvzInfo, error)
	CreateReception(ctx context.Context, pvzID string, dateTime time.Time) (string, error)
	AddProduct(ctx context.Context, pvzID string, productType string) (string, error)
	DeleteLastProduct(ctx context.Context, pvzID string) error
	CloseReception(ctx context.Context, pvzID string) (string, error)
	GetPvzsInfoOptimized(ctx context.Context, page, limit int, startDate, endDate *time.Time) ([]entity.PvzInfo, error)
}

type pvzServiceImp struct {
	repo                db.Repository
	txManager           db.TxManager
	logger              logger.Logger
	allowedCities       map[string]bool
	allowedProductTypes map[string]bool
}

func NewPvzService(repo db.Repository, txManager db.TxManager, logger logger.Logger, cities, productTypes map[string]bool) PvzService {
	return &pvzServiceImp{
		repo:                repo,
		txManager:           txManager,
		logger:              logger,
		allowedCities:       cities,
		allowedProductTypes: productTypes,
	}
}

func (s *pvzServiceImp) CreatePvz(ctx context.Context, city string) (string, error) {
	ctx, span := otel.Tracer("pvzService").Start(ctx, "CreatePvzService")
	defer span.End()

	span.SetAttributes(attribute.String("city", city))

	if !s.allowedCities[strings.ToLower(city)] {
		return "", errs.New(errs.ErrInvalidCity, fmt.Sprintf("city '%s' is not allowed", city))
	}

	pvz := entity.Pvz{
		City:             city,
		RegistrationDate: time.Now(),
	}

	pvzID, err := s.repo.CreatePvz(ctx, pvz)
	if err != nil {
		s.logger.Errorw("CreatePvz",
			"error", err,
			"city", city,
		)
		return "", err
	}
	metrics.PVZCreated()
	return pvzID, nil
}

func (s *pvzServiceImp) GetPvzsInfo(ctx context.Context, page, limit int, startDate, endDate *time.Time) ([]entity.PvzInfo, error) {
	var pvzsInfo []entity.PvzInfo

	err := s.txManager.WithTx(ctx, pgx.ReadCommitted, pgx.ReadOnly, func(txCtx context.Context) error {
		pvzs, err := s.repo.GetPvzs(txCtx, page, limit)
		if err != nil {
			s.logger.Errorw("get PVZs",
				"error", err,
				"page", page,
				"limit", limit,
			)
			return errs.Wrap(err, errs.ErrInternalCode, "failed to get pvzs")
		}

		for _, pvz := range pvzs {
			receptions, err := s.repo.GetReceptionsByPvzIDFiltered(txCtx, pvz.ID, startDate, endDate)
			if err != nil {
				s.logger.Errorw("get receptions",
					"error", err,
					"pvzID", pvz.ID,
				)
				return errs.Wrap(err, errs.ErrInternalCode, fmt.Sprintf("failed to get receptions for pvz %s", pvz.ID))
			}

			var recInfos []entity.ReceptionInfo
			for _, rec := range receptions {
				products, err := s.repo.GetProductsByReceptionID(txCtx, rec.ID)
				if err != nil {
					s.logger.Errorw("get products",
						"error", err,
						"receptionID", rec.ID,
					)
					return errs.Wrap(err, errs.ErrInternalCode, fmt.Sprintf("failed to get products for reception %s", rec.ID))
				}

				recInfos = append(recInfos, entity.ReceptionInfo{
					Reception: rec,
					Products:  products,
				})
			}

			pvzInfo := entity.PvzInfo{
				Pvz:        pvz,
				Receptions: recInfos,
			}
			pvzsInfo = append(pvzsInfo, pvzInfo)
		}
		return nil
	})
	if err != nil {
		s.logger.Errorw("GetPvzsInfo transaction failed",
			"error", err,
		)
		return nil, err
	}

	return pvzsInfo, nil
}

func (s *pvzServiceImp) CreateReception(ctx context.Context, pvzID string, dateTime time.Time) (string, error) {
	var receptionID string
	err := s.txManager.WithTx(ctx, pgx.ReadCommitted, pgx.ReadWrite, func(txCtx context.Context) error {
		existingReception, err := s.repo.FindOpenReceptionByPvzID(txCtx, pvzID)
		if err == nil && existingReception != nil {
			return errs.New(errs.ErrOpenReceptionExists, "open reception already exists")
		}

		if err != nil && !errs.IsOpenReceptionNotFound(err) {
			return err
		}

		rec := entity.Reception{
			PvzID:    pvzID,
			DateTime: dateTime,
			Status:   "in_progress",
		}

		id, err := s.repo.CreateReception(txCtx, rec)
		if err != nil {
			return err
		}
		receptionID = id
		return nil
	})
	if err != nil {
		s.logger.Errorw("CreateReception",
			"error", err,
			"pvzID", pvzID,
		)
		return "", err
	}

	metrics.ReceptionsCreated()
	return receptionID, nil
}

func (s *pvzServiceImp) AddProduct(ctx context.Context, pvzID string, productType string) (string, error) {
	if !s.allowedProductTypes[strings.ToLower(productType)] {
		return "", errs.New(errs.ErrInvalidProductType, fmt.Sprintf("product type '%s' is not allowed", productType))
	}

	var productID string
	err := s.txManager.WithTx(ctx, pgx.ReadCommitted, pgx.ReadWrite, func(txCtx context.Context) error {
		reception, err := s.repo.FindOpenReceptionByPvzID(txCtx, pvzID)
		if err != nil {
			return err
		}
		if reception == nil {
			return errs.New(errs.ErrNoOpenReception, "no open reception found for this PVZ")
		}

		product := entity.Product{
			ReceptionID: reception.ID,
			Type:        productType,
			DateTime:    time.Now(),
		}
		id, err := s.repo.CreateProduct(txCtx, product)
		if err != nil {
			return err
		}
		productID = id
		return nil
	})
	if err != nil {
		s.logger.Errorw("AddProduct",
			"error", err,
			"pvzID", pvzID,
			"productType", productType,
		)
		return "", err
	}

	metrics.ProductsAdded()
	return productID, nil
}

func (s *pvzServiceImp) DeleteLastProduct(ctx context.Context, pvzID string) error {
	err := s.txManager.WithTx(ctx, pgx.ReadCommitted, pgx.ReadWrite, func(txCtx context.Context) error {
		product, err := s.repo.FindLastProductInReception(txCtx, pvzID)
		if err != nil {
			return err
		}
		if product == nil {
			return errs.New(errs.ErrNoProductsToDelete, "no product found to delete")
		}
		return s.repo.DeleteProduct(txCtx, product.ID)
	})
	if err != nil {
		s.logger.Errorw("DeleteLastProduct",
			"error", err,
			"pvzID", pvzID,
		)
		return err
	}
	return nil
}

func (s *pvzServiceImp) CloseReception(ctx context.Context, pvzID string) (string, error) {
	var recID string
	err := s.txManager.WithTx(ctx, pgx.ReadCommitted, pgx.ReadWrite, func(txCtx context.Context) error {
		reception, err := s.repo.FindOpenReceptionByPvzID(txCtx, pvzID)
		if err != nil {
			return err
		}
		if reception == nil {
			return errs.New(errs.ErrReceptionNotFound, "no open reception found for this PVZ")
		}
		if err := s.repo.UpdateReceptionStatus(txCtx, reception.ID, "close"); err != nil {
			return err
		}
		recID = reception.ID
		return nil
	})
	if err != nil {
		s.logger.Errorw("CloseReception",
			"error", err,
			"pvzID", pvzID,
		)
		return "", err
	}
	return recID, nil
}

func (s *pvzServiceImp) GetPvzsInfoOptimized(
	ctx context.Context,
	page, limit int,
	startDate, endDate *time.Time,
) ([]entity.PvzInfo, error) {
	pvzs, err := s.repo.GetPvzsWithReceptionsAndProducts(ctx, page, limit, startDate, endDate)
	if err != nil {
		s.logger.Errorw("GetPvzsInfoOptimized failed",
			"error", err,
			"page", page,
			"limit", limit,
		)
		return nil, err
	}
	return pvzs, nil
}
