package main

import (
	"mqtt-bench/config"
	"mqtt-bench/generator"
	"mqtt-bench/logger"
	"mqtt-bench/metrics"
	"mqtt-bench/publisher"
	"mqtt-bench/report"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Message struct {
	Topic   string
	Payload int
}

func main() {
	startTime := time.Now()

	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// // Инициализация HTTP-сервера
	met := metrics.NewMetrics()
	met.Init(cfg.PrometheusPort)

	// Инициализация логгера
	logger.Init()

	// Настройка MQTT клиента
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.BrokerURL).
		SetClientID(cfg.ClientIDPrefix)

	// Аутентификация если указаны учетные данные
	if cfg.MQTTUsername != "" {
		opts.SetUsername(cfg.MQTTUsername)
		if cfg.MQTTPassword != "" {
			opts.SetPassword(cfg.MQTTPassword)
		}
	}

	// Инициализация клиента MQTT
	client := mqtt.NewClient(opts)

	// Подключение к брокеру
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logger.Log.Fatal().Err(token.Error()).Msgf("Failed to connect to MQTT broker %s", cfg.BrokerURL)
	}
	logger.Log.Info().Msgf("Succesful to connect to MQTT broker %s", cfg.BrokerURL)

	// Генерирурем в ОЗУ TotalSignal
	messages := generator.GenMessages(cfg)

	// Публикуем значения в заданном режиме
	publisher.ModeRule(client, cfg, messages, met)

	// Генерация отчёта
	endTime := time.Now()
	err := report.WriteReport(cfg, met, startTime, endTime)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to write report")
	} else {
		logger.Log.Info().Msgf("Report written to %s", cfg.ReportFile)
	}

	logger.Log.Info().Msgf("Benchmark finished ant total published messages = %d", met.GetSentCount())

}
