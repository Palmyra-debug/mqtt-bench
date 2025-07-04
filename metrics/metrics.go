package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/atomic"
)

type Metrics struct {
	registry *prometheus.Registry

	// Prometheus метрики
	sentTotal       prometheus.Counter
	failedTotal     prometheus.Counter
	publishLatency  prometheus.Histogram
	throughputGauge prometheus.Gauge
	bytesPayload    prometheus.Counter
	bytesTopic      prometheus.Counter
	uptimeGauge     prometheus.Gauge

	// Локальные атомарные счётчики
	sentCount  atomic.Uint64
	errorCount atomic.Uint64

	// Время старта клиента
	startTime time.Time
}

// NewMetrics создаёт новый экземпляр Metrics с зарегистрированными метриками
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		registry: reg,

		sentTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "messages_sent_total",
			Help: "Total number of successfully published messages",
		}),
		failedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "messages_failed_total",
			Help: "Total number of failed message publishes",
		}),
		publishLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "publish_latency_seconds",
			Help:    "Histogram of MQTT publish latency",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 15),
		}),
		throughputGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "throughput_msgs_sec",
			Help: "Number of MQTT messages published per second",
		}),
		bytesPayload: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "bytes_sent_payload",
			Help: "Total bytes sent as payload",
		}),
		bytesTopic: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "bytes_sent_topic",
			Help: "Total bytes sent as topic strings",
		}),
		uptimeGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "uptime_seconds",
			Help: "Uptime of the MQTT client in seconds",
		}),
		startTime: time.Now(),
	}

	// Регистрируем метрики в отдельном реестре
	reg.MustRegister(
		m.sentTotal,
		m.failedTotal,
		m.publishLatency,
		m.throughputGauge,
		m.bytesPayload,
		m.bytesTopic,
		m.uptimeGauge,
	)

	// Запускаем горутину для обновления uptime каждую секунду
	go func() {
		for {
			time.Sleep(time.Second)
			m.uptimeGauge.Set(time.Since(m.startTime).Seconds())
		}
	}()

	return m
}

// Init запускает HTTP-сервер с двумя эндпоинтами на одном порту:
// - /metrics — твои кастомные метрики из кастомного реестра
// - /debug/metrics — системные метрики из дефолтного реестра Prometheus
func (m *Metrics) Init(port int) chan error {
	errCh := make(chan error, 1) // буферизованный канал, чтобы не блокировать

	go func() {
		addr := ":" + strconv.Itoa(port)
		mux := http.NewServeMux()

		// Кастомные метрики
		mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))

		// Стандартные системные метрики
		mux.Handle("/debug/metrics", promhttp.Handler())

		err := http.ListenAndServe(addr, mux)
		errCh <- err // отправляем ошибку, если она будет
	}()

	return errCh
}

func (m *Metrics) IncSent(payloadSize int, topicSize int) {
	m.sentTotal.Inc()
	m.sentCount.Inc()
	m.bytesPayload.Add(float64(payloadSize))
	m.bytesTopic.Add(float64(topicSize))
}

func (m *Metrics) IncError() {
	m.failedTotal.Inc()
	m.errorCount.Inc()
}

// Измерение латентности
func (m *Metrics) ObserveLatency(d time.Duration) {
	m.publishLatency.Observe(d.Seconds())
}

func (m *Metrics) SetThroughput(msgsPerSecond float64) {
	m.throughputGauge.Set(msgsPerSecond)
}

func (m *Metrics) GetSentCount() uint64 {
	return m.sentCount.Load()
}

func (m *Metrics) GetErrorCount() uint64 {
	return m.errorCount.Load()
}
