package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// PVZCreatedTotal — счетчик количества созданных ПВЗ
	PVZCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Total number of created PVZs.",
		},
	)

	// ReceptionsCreatedTotal — счетчик количества созданных приёмок заказов
	ReceptionsCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "receptions_created_total",
			Help: "Total number of created receptions.",
		},
	)

	// ProductsAddedTotal — счетчик количества добавленных товаров
	ProductsAddedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "products_added_total",
			Help: "Total number of added products.",
		},
	)
)

func init() {
	prometheus.MustRegister(PVZCreatedTotal, ReceptionsCreatedTotal, ProductsAddedTotal)
}

func PVZCreated() {
	PVZCreatedTotal.Inc()
}

func ReceptionsCreated() {
	ReceptionsCreatedTotal.Inc()
}

func ProductsAdded() {
	ProductsAddedTotal.Inc()
}
