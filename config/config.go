package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Основные параметры
	BrokerURL      string
	MQTTUsername   string
	MQTTPassword   string
	ClientIDPrefix string
	TopicPattern   string
	QoS            int
	Mode           string

	// Общие параметры
	TotalSignals   int
	PrometheusPort int
	ReportFile     string
	LogLevel       string

	// Режим constant
	SignalsPerSecond int

	// Режим burst
	BurstSize       int
	BurstIntervalMs int

	// Режим ramp-up
	StartRate       int
	EndRate         int
	RampDurationSec int

	// Режим random
	MinRate int
	MaxRate int
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	getEnv := func(key string, defaultVal string) string {
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		log.Printf("- WARN - Missing parameter %s in the configuration. The default value is accepted", key)
		return defaultVal
	}

	getInt := func(key string, defaultVal int) int {
		valStr := getEnv(key, "")
		if valStr == "" {
			log.Printf("- WARN - Missing parameter %s in the configuration. The default value is accepted", key)
			return defaultVal
		}
		// Преобразование строки в число
		val, err := strconv.Atoi(valStr)
		if err != nil {
			// Тип ошибки который важно предотвратить до запуска приложения
			log.Fatalf("Invalid int for %s: %v", key, err)
		}
		return val
	}

	return &Config{
		// Основные параметры
		BrokerURL:      getEnv("BROKER_URL", "tcp://localhost:1883"),
		ClientIDPrefix: getEnv("CLIENT_ID_PREFIX", "benchClient"),
		MQTTUsername:   getEnv("MQTT_USERNAME", ""),
		MQTTPassword:   getEnv("MQTT_PASSWORD", ""),
		TopicPattern:   getEnv("TOPIC_PATTERN", "/devices/{device_id}/controls/{control_id}"),
		QoS:            getInt("QoS", 0),
		Mode:           getEnv("MODE", "constant"),

		// Общие параметры
		TotalSignals:   getInt("TOTAL_SIGNALS", 1000),
		PrometheusPort: getInt("PROMETHEUS_PORT", 2112),
		ReportFile:     getEnv("REPORT_FILE", "report.json"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),

		// Режим constant
		SignalsPerSecond: getInt("SIGNALS_PER_SECOND", 100),

		// Режим burst
		BurstSize:       getInt("BURST_SIZE", 100),
		BurstIntervalMs: getInt("BURST_INTERVAL_MS", 1000),

		// Режим ramp-up
		StartRate:       getInt("START_RATE", 50),
		EndRate:         getInt("END_RATE", 1000),
		RampDurationSec: getInt("RAMP_DURATION_SEC", 300),

		// Режим random
		MinRate: getInt("MIN_RATE", 100),
		MaxRate: getInt("MAX_RATE", 1000),
	}
}
