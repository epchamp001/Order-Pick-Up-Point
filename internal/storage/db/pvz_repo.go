package db

import (
	"context"
	"order-pick-up-point/internal/errs"
	"order-pick-up-point/internal/metrics"
	"order-pick-up-point/internal/models/entity"
	"order-pick-up-point/pkg/logger"
	"time"
)

type PvzRepository interface {
	CreatePvz(ctx context.Context, pvz entity.Pvz) (string, error)
	GetPvzs(ctx context.Context, page, limit int) ([]entity.Pvz, error)
	GetListOfPvzs(ctx context.Context) ([]entity.Pvz, error)
}

type postgresPvzRepository struct {
	conn   TxManager
	logger logger.Logger
}

func NewPvzRepository(conn TxManager, log logger.Logger) PvzRepository {
	return &postgresPvzRepository{
		conn:   conn,
		logger: log,
	}
}

func (r *postgresPvzRepository) CreatePvz(ctx context.Context, pvz entity.Pvz) (string, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("CreatePvz", time.Since(start).Seconds())
	}()

	pool := r.conn.GetExecutor(ctx)
	query := `
		INSERT INTO pvz (registration_date, city)
		VALUES ($1, $2)
		RETURNING id
	`

	var pvzID string
	err := pool.QueryRow(ctx, query, pvz.RegistrationDate, pvz.City).Scan(&pvzID)
	if err != nil {
		r.logger.Errorw("creating PVZ",
			"error", err,
			"city", pvz.City,
		)
		return "", errs.Wrap(err, errs.ErrInternalCode, "failed to create pvz")
	}
	return pvzID, nil
}

func (r *postgresPvzRepository) GetPvzs(ctx context.Context, page, limit int) ([]entity.Pvz, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("GetPvzs", time.Since(start).Seconds())
	}()

	offset := (page - 1) * limit
	query := `
		SELECT id, registration_date, city
		FROM pvz
		ORDER BY registration_date DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.conn.GetExecutor(ctx).Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.Errorw("getting PVZs",
			"error", err,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to get pvzs")
	}
	defer rows.Close()

	var pvzs []entity.Pvz
	for rows.Next() {
		var p entity.Pvz
		if err := rows.Scan(&p.ID, &p.RegistrationDate, &p.City); err != nil {
			r.logger.Errorw("scanning PVZ",
				"error", err,
			)
			return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to scan pvz")
		}
		pvzs = append(pvzs, p)
	}
	if err = rows.Err(); err != nil {
		r.logger.Errorw("Rows error in GetPvzs",
			"error", err,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "rows error")
	}
	return pvzs, nil
}

func (r *postgresPvzRepository) GetListOfPvzs(ctx context.Context) ([]entity.Pvz, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("GetListOfPvzs", time.Since(start).Seconds())
	}()

	query := `
		SELECT id, registration_date, city
		FROM pvz
		ORDER BY registration_date DESC
	`
	rows, err := r.conn.GetExecutor(ctx).Query(ctx, query)
	if err != nil {
		r.logger.Errorw("query error",
			"error", err,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to get PVZs")
	}
	defer rows.Close()

	var pvzs []entity.Pvz
	for rows.Next() {
		var pvz entity.Pvz
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
			r.logger.Errorw("scan error",
				"error", err,
			)
			return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to scan PVZ")
		}
		pvzs = append(pvzs, pvz)
	}
	if err = rows.Err(); err != nil {
		r.logger.Errorw("rows error",
			"error", err,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "rows iteration error")
	}
	return pvzs, nil
}
