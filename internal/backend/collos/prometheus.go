package collos

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ad = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "backend_collos_api_duration_seconds",
		Help: "The duration of Collos API calls (per endpoint).",
	}, []string{"endpoint"})
)

func collosAPIDuration(e string) prometheus.Observer {
	return ad.With(prometheus.Labels{"endpoint": e})
}
