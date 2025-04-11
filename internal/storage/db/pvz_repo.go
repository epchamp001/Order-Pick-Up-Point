package db

import (
	"context"
	"fmt"
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
	GetPvzsWithReceptionsAndProducts(ctx context.Context, page, limit int, startDate, endDate *time.Time) ([]entity.PvzInfo, error)
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

func (r *postgresPvzRepository) GetPvzsWithReceptionsAndProducts(
	ctx context.Context,
	page, limit int,
	startDate, endDate *time.Time,
) ([]entity.PvzInfo, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("GetPvzsWithReceptionsAndProducts", time.Since(start).Seconds())
	}()

	offset := (page - 1) * limit
	args := []interface{}{limit, offset}
	argIndex := 3

	query := `
		SELECT
			p.id AS pvz_id,
			p.registration_date,
			p.city,
			r.id AS reception_id,
			r.date_time AS reception_date,
			r.status,
			pr.id AS product_id,
			pr.date_time AS product_date,
			pr.type
		FROM pvz p
		LEFT JOIN reception r ON r.pvz_id = p.id
		LEFT JOIN product pr ON pr.reception_id = r.id
	`

	if startDate != nil && endDate != nil {
		query += fmt.Sprintf(" WHERE r.date_time BETWEEN $%d AND $%d", argIndex, argIndex+1)
		args = append(args, startDate, endDate)
		argIndex += 2
	}

	query += " ORDER BY p.registration_date DESC LIMIT $1 OFFSET $2"

	rows, err := r.conn.GetExecutor(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to get pvz with receptions and products")
	}
	defer rows.Close()

	pvzMap := make(map[string]*entity.PvzInfo)

	for rows.Next() {
		var (
			pvzID, city, receptionID, status, productID, productType *string
			pvzRegDate, receptionDate, productDate                   *time.Time
		)

		if err := rows.Scan(&pvzID, &pvzRegDate, &city, &receptionID, &receptionDate, &status, &productID, &productDate, &productType); err != nil {
			return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to scan pvz row")
		}

		if _, exists := pvzMap[*pvzID]; !exists {
			pvzMap[*pvzID] = &entity.PvzInfo{
				Pvz: entity.Pvz{
					ID:               *pvzID,
					RegistrationDate: *pvzRegDate,
					City:             *city,
				},
			}
		}

		pvzInfo := pvzMap[*pvzID]

		if receptionID != nil {
			receptionExists := false
			for i := range pvzInfo.Receptions {
				if pvzInfo.Receptions[i].Reception.ID == *receptionID {
					receptionExists = true
					if productID != nil {
						pvzInfo.Receptions[i].Products = append(
							pvzInfo.Receptions[i].Products,
							entity.Product{
								ID:          *productID,
								DateTime:    *productDate,
								Type:        *productType,
								ReceptionID: *receptionID,
							},
						)
					}
					break
				}
			}
			if !receptionExists {
				newReception := entity.ReceptionInfo{
					Reception: entity.Reception{
						ID:       *receptionID,
						DateTime: *receptionDate,
						PvzID:    *pvzID,
						Status:   *status,
					},
				}
				if productID != nil {
					newReception.Products = append(newReception.Products, entity.Product{
						ID:          *productID,
						DateTime:    *productDate,
						Type:        *productType,
						ReceptionID: *receptionID,
					})
				}
				pvzInfo.Receptions = append(pvzInfo.Receptions, newReception)
			}
		}
	}

	var result []entity.PvzInfo
	for _, info := range pvzMap {
		result = append(result, *info)
	}

	return result, nil
}
