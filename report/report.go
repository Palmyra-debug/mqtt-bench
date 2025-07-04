package report

import (
	"encoding/json"
	"os"
	"time"

	"mqtt-bench/config"
	"mqtt-bench/metrics"
)

type Report struct {
	Mode         string    `json:"mode"`
	TotalSignals int       `json:"total_signals"`
	SuccessCount uint64    `json:"success_count"`
	ErrorCount   uint64    `json:"error_count"`
	DurationSec  float64   `json:"duration_sec"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	AvgRate      float64   `json:"avg_rate_msgs_per_sec"`
}

func WriteReport(cfg *config.Config, met *metrics.Metrics, startTime time.Time, endTime time.Time) error {
	success := met.GetSentCount()
	errors := met.GetErrorCount()
	duration := endTime.Sub(startTime).Seconds()

	var avgRate float64
	if duration > 0 {
		avgRate = float64(success) / duration
	}

	report := Report{
		Mode:         cfg.Mode,
		TotalSignals: cfg.TotalSignals,
		SuccessCount: success,
		ErrorCount:   errors,
		DurationSec:  duration,
		StartTime:    startTime,
		EndTime:      endTime,
		AvgRate:      avgRate,
	}

	file, err := os.Create(cfg.ReportFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
