package db

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"order-pick-up-point/internal/errs"
	"order-pick-up-point/internal/metrics"
	"order-pick-up-point/internal/models/entity"
	"order-pick-up-point/pkg/logger"
	"time"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product entity.Product) (string, error)
	FindLastProductInReception(ctx context.Context, pvzID string) (*entity.Product, error)
	DeleteProduct(ctx context.Context, productID string) error
	GetProductsByReceptionID(ctx context.Context, receptionID string) ([]entity.Product, error)
}

type postgresProductRepository struct {
	conn   TxManager
	logger logger.Logger
}

func NewProductRepository(conn TxManager, log logger.Logger) ProductRepository {
	return &postgresProductRepository{
		conn:   conn,
		logger: log,
	}
}

func (r *postgresProductRepository) CreateProduct(ctx context.Context, product entity.Product) (string, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("CreateProduct", time.Since(start).Seconds())
	}()
	pool := r.conn.GetExecutor(ctx)
	query := `
		INSERT INTO product (date_time, type, reception_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var productID string
	err := pool.QueryRow(ctx, query, product.DateTime, product.Type, product.ReceptionID).Scan(&productID)
	if err != nil {
		r.logger.Errorw("creating product",
			"error", err,
			"productType", product.Type,
			"receptionID", product.ReceptionID,
		)
		return "", errs.Wrap(err, errs.ErrInternalCode, "failed to create product")
	}
	return productID, nil
}

func (r *postgresProductRepository) FindLastProductInReception(ctx context.Context, pvzID string) (*entity.Product, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("FindLastProductInReception", time.Since(start).Seconds())
	}()
	pool := r.conn.GetExecutor(ctx)

	// Выполняем JOIN с таблицей reception для выбора последнего товара из открытой приёмки данного ПВЗ
	query := `
		SELECT p.id, p.date_time, p.type, p.reception_id
		FROM product p
		JOIN reception r ON p.reception_id = r.id
		WHERE r.pvz_id = $1 AND r.status = 'in_progress'
		ORDER BY p.date_time DESC
		LIMIT 1
	`

	var prod entity.Product
	err := pool.QueryRow(ctx, query, pvzID).Scan(
		&prod.ID,
		&prod.DateTime,
		&prod.Type,
		&prod.ReceptionID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.New(errs.ErrNoProductsToDelete, "no product found to delete")
		}
		r.logger.Errorw("finding last product in reception",
			"error", err,
			"pvzID", pvzID,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to find last product in reception")
	}
	return &prod, nil
}

func (r *postgresProductRepository) DeleteProduct(ctx context.Context, productID string) error {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("DeleteProduct", time.Since(start).Seconds())
	}()

	pool := r.conn.GetExecutor(ctx)
	query := `
		DELETE FROM product
		WHERE id = $1
	`
	cmdTag, err := pool.Exec(ctx, query, productID)
	if err != nil {
		r.logger.Errorw("deleting product",
			"error", err,
			"productID", productID,
		)
		return errs.Wrap(err, errs.ErrInternalCode, "failed to delete product")
	}
	if cmdTag.RowsAffected() == 0 {
		return errs.New(errs.ErrNoProductsToDelete, "no product found with provided id")
	}
	return nil
}

func (r *postgresProductRepository) GetProductsByReceptionID(ctx context.Context, receptionID string) ([]entity.Product, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("GetProductsByReceptionID", time.Since(start).Seconds())
	}()

	query := `
		SELECT id, date_time, type, reception_id
		FROM product
		WHERE reception_id = $1
		ORDER BY date_time
	`
	rows, err := r.conn.GetExecutor(ctx).Query(ctx, query, receptionID)
	if err != nil {
		r.logger.Errorw("query error",
			"error", err,
			"receptionID", receptionID,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to get products")
	}
	defer rows.Close()

	var products []entity.Product
	for rows.Next() {
		var prod entity.Product
		if err := rows.Scan(&prod.ID, &prod.DateTime, &prod.Type, &prod.ReceptionID); err != nil {
			r.logger.Errorw("scan error",
				"error", err,
				"receptionID", receptionID,
			)
			return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to scan product")
		}
		products = append(products, prod)
	}
	if err = rows.Err(); err != nil {
		r.logger.Errorw("rows error",
			"error", err,
			"receptionID", receptionID,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "rows iteration error")
	}
	return products, nil
}
