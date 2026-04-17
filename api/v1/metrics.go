package v1

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ActivePlayers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "golink_active_players",
		Help: "Number of active players (connected to voice)",
	})
	PlayingPlayers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "golink_playing_players",
		Help: "Number of players currently playing audio",
	})
	TracksLoaded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "golink_tracks_loaded_total",
		Help: "Total tracks loaded by source",
	}, []string{"source"})
	TrackLoadDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "golink_track_load_duration_seconds",
		Help:    "Track resolution duration",
		Buckets: prometheus.DefBuckets,
	}, []string{"source"})
	WebSocketConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "golink_websocket_connections",
		Help: "Number of active WebSocket client connections",
	})
)

func RegisterMetrics() {
	prometheus.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		ActivePlayers,
		PlayingPlayers,
		TracksLoaded,
		TrackLoadDuration,
		WebSocketConnections,
	)
}

func MetricsHandler() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.Handler())
}
