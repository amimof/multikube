package multikube

import (
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricsReqsTotal = "multikube_requests_total"
)

type metrics struct {
	counters map[string]prometheus.Collector
}

func initMetrics() *metrics {
	metrics := &metrics{
		counters: make(map[string]prometheus.Collector),
	}
	numreqs := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricsReqsTotal,
			Help: "How many HTTP requests processed on an entrypoint, partitioned by status code, protocol, and method.",
		}, 
		[]string{"code", "method", "protocol"},
	)
	metrics.register(MetricsReqsTotal, numreqs)
	return metrics
}

func (m *metrics) register(name string, c prometheus.Collector) {
	m.counters[name] = c
	prometheus.MustRegister(c)
}

func (m *metrics) getCounter(req *http.Request, name string) prometheus.Counter {
	var c prometheus.Counter
	cv := m.getCounterVec(name)
	if cv != nil {
		c = cv.With(prometheus.Labels{
			"code": string(req.Response.StatusCode),
			"method": req.Method,
			"protocol": req.Proto,
		})
	}
	return c
}

func (m *metrics) getCounterVec(name string) *prometheus.CounterVec {
	for k, _ := range m.counters {
		if k == name {
			if v, ok := m.counters[k].(*prometheus.CounterVec); ok {
				return v
			}
		}
	}
	return nil
}