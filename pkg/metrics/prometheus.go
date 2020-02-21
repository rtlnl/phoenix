package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus struct to fulfil the metrics interface
type Prometheus struct {
	RecommendRequests *prometheus.CounterVec
	RecommendLatency  prometheus.Summary
	Timer             *prometheus.Timer
}

// NewPrometheus instantiates a new prometheus client object
func NewPrometheus() *Prometheus {
	rr := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "phoenix",
			Subsystem: "public",
			Name:      "recommend_requests_total",
			Help:      "How many Recommend requests processed, partitioned by status",
		},
		[]string{"status"},
	)

	rl := prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:  "phoenix",
			Subsystem:  "public",
			Name:       "recommend_request_durations",
			Help:       "Recommend requests latencies in milliseconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		})

	prometheus.MustRegister(rr)
	prometheus.MustRegister(rl)

	return &Prometheus{RecommendRequests: rr, RecommendLatency: rl}
}

func (p *Prometheus) FailedRequest() {
	p.RecommendRequests.WithLabelValues("fail").Inc()
}

func (p *Prometheus) SuccessRequest() {
	p.RecommendRequests.WithLabelValues("success").Inc()
}

func (p *Prometheus) NotFoundRequest() {
	p.RecommendRequests.WithLabelValues("not_found").Inc()
}

func (p *Prometheus) StartTimer() {
	p.Timer = prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000
		p.RecommendLatency.Observe(us)
	}))
}

func (p *Prometheus) Latency() {
	p.Timer.ObserveDuration()
}
