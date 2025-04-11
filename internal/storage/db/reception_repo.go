package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"order-pick-up-point/internal/errs"
	"order-pick-up-point/internal/metrics"
	"order-pick-up-point/internal/models/entity"
	"order-pick-up-point/pkg/logger"
	"time"
)

type ReceptionRepository interface {
	FindOpenReceptionByPvzID(ctx context.Context, pvzID string) (*entity.Reception, error)
	CreateReception(ctx context.Context, reception entity.Reception) (string, error)
	UpdateReceptionStatus(ctx context.Context, receptionID string, status string) error
	GetReceptionsByPvzIDFiltered(ctx context.Context, pvzID string, startDate, endDate *time.Time) ([]entity.Reception, error)
}

type postgresReceptionRepository struct {
	conn   TxManager
	logger logger.Logger
}

func NewReceptionRepository(conn TxManager, log logger.Logger) ReceptionRepository {
	return &postgresReceptionRepository{conn: conn, logger: log}
}

func (r *postgresReceptionRepository) FindOpenReceptionByPvzID(ctx context.Context, pvzID string) (*entity.Reception, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("FindOpenReceptionByPvzID", time.Since(start).Seconds())
	}()

	pool := r.conn.GetExecutor(ctx)
	query := `
		SELECT id, date_time, pvz_id, status
		FROM reception
		WHERE pvz_id = $1 AND status = 'in_progress'
		ORDER BY date_time DESC
		LIMIT 1
	`
	var rec entity.Reception
	err := pool.QueryRow(ctx, query, pvzID).Scan(&rec.ID, &rec.DateTime, &rec.PvzID, &rec.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.New(errs.ErrNoOpenReception, "open reception not found")
		}
		r.logger.Errorw("finding open reception by pvz",
			"error", err,
			"pvzID", pvzID,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to find open reception by pvz id")
	}
	return &rec, nil
}

func (r *postgresReceptionRepository) CreateReception(ctx context.Context, reception entity.Reception) (string, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("CreateReception", time.Since(start).Seconds())
	}()

	pool := r.conn.GetExecutor(ctx)
	query := `
		INSERT INTO reception (date_time, pvz_id, status)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var receptionID string
	err := pool.QueryRow(ctx, query, reception.DateTime, reception.PvzID, reception.Status).Scan(&receptionID)
	if err != nil {
		r.logger.Errorw("creating reception",
			"error", err,
			"pvzID", reception.PvzID,
		)
		return "", errs.Wrap(err, errs.ErrInternalCode, "failed to create reception")
	}
	return receptionID, nil
}

func (r *postgresReceptionRepository) UpdateReceptionStatus(ctx context.Context, receptionID string, status string) error {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("UpdateReceptionStatus", time.Since(start).Seconds())
	}()

	pool := r.conn.GetExecutor(ctx)
	query := `
		UPDATE reception
		SET status = $1
		WHERE id = $2
	`
	cmdTag, err := pool.Exec(ctx, query, status, receptionID)
	if err != nil {
		r.logger.Errorw("updating reception status",
			"error", err,
			"receptionID", receptionID,
			"status", status,
		)
		return errs.Wrap(err, errs.ErrInternalCode, "failed to update reception status")
	}

	if cmdTag.RowsAffected() == 0 {
		return errs.New(errs.ErrReceptionNotFound, "no reception found to update")
	}
	return nil
}

func (r *postgresReceptionRepository) GetReceptionsByPvzIDFiltered(ctx context.Context, pvzID string, startDate, endDate *time.Time) ([]entity.Reception, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("GetReceptionsByPvzIDFiltered", time.Since(start).Seconds())
	}()

	query := `
		SELECT id, date_time, pvz_id, status 
		FROM reception 
		WHERE pvz_id = $1
	`

	args := []interface{}{pvzID}
	argIndex := 2

	if startDate != nil && endDate != nil {
		query += fmt.Sprintf(" AND date_time BETWEEN $%d AND $%d", argIndex, argIndex+1)
		args = append(args, startDate, endDate)
		argIndex += 2
	}
	query += " ORDER BY date_time"

	rows, err := r.conn.GetExecutor(ctx).Query(ctx, query, args...)
	if err != nil {
		r.logger.Errorw("query error",
			"error", err,
			"pvzID", pvzID,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to get receptions")
	}
	defer rows.Close()

	var receptions []entity.Reception
	for rows.Next() {
		var rec entity.Reception
		if err := rows.Scan(&rec.ID, &rec.DateTime, &rec.PvzID, &rec.Status); err != nil {
			r.logger.Errorw("scan error",
				"error", err,
				"pvzID", pvzID,
			)
			return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to scan reception")
		}
		receptions = append(receptions, rec)
	}
	if err := rows.Err(); err != nil {
		r.logger.Errorw("rows error",
			"error", err,
			"pvzID", pvzID,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "rows iteration error")
	}
	return receptions, nil
}
