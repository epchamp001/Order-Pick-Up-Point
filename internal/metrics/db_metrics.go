package metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var (
	// DBQueryDuration - гистограмма времени выполнения запросов к базе данных
	DBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query"},
	)

	// DBActiveConnections - gauge для количества активных соединений с базой данных
	DBActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_active_connections",
			Help: "Number of active database connections.",
		},
	)
)

func init() {
	prometheus.MustRegister(DBQueryDuration, DBActiveConnections)
}

func RecordDBQueryDuration(query string, duration float64) {
	DBQueryDuration.WithLabelValues(query).Observe(duration)
}

func RecordDBActiveConnections(count int) {
	DBActiveConnections.Set(float64(count))
}

func MonitorDBConnections(pgPool *pgxpool.Pool, interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				stats := pgPool.Stat()
				RecordDBActiveConnections(int(stats.TotalConns()))
			case <-stop:
				return
			}
		}
	}()
}
