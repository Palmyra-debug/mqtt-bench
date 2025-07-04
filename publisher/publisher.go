package publisher

import (
	"mqtt-bench/config"
	"mqtt-bench/generator"

	"math/rand"
	"mqtt-bench/logger"
	"mqtt-bench/metrics" // обязательно импортируй свой пакет с метриками
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func ModeRule(client mqtt.Client, cfg *config.Config, messages []generator.Message, met *metrics.Metrics) {
	switch cfg.Mode {
	case "constant":
		ConstantMode(client, cfg, messages, met)
	case "burst":
		BurstMode(client, cfg, messages, met)
	case "ramp-up":
		RampUpMode(client, cfg, messages, met)
	case "random":
		RandomMode(client, cfg, messages, met)
	default:
		logger.Log.Fatal().Msgf("Unknown mode: %s", cfg.Mode)
	}
}

func ConstantMode(client mqtt.Client, cfg *config.Config, messages []generator.Message, met *metrics.Metrics) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	totalSteps := (len(messages) + cfg.SignalsPerSecond - 1) / cfg.SignalsPerSecond
	currentStep := 0
	lastBatchTime := time.Now()

	for currentStep < totalSteps {
		<-ticker.C

		startIdx := currentStep * cfg.SignalsPerSecond
		endIdx := startIdx + cfg.SignalsPerSecond
		if endIdx > len(messages) {
			endIdx = len(messages)
		}
		// Определение размера пачки и дальнейший расчет средней пропускной способности одной публикации в разрезе пачки
		now := time.Now()
		elapsed := now.Sub(lastBatchTime).Seconds()
		if elapsed > 0 {
			batchSize := float64(endIdx - startIdx)
			met.SetThroughput(batchSize / elapsed)
		}
		lastBatchTime = now

		currentStep++
		for _, msg := range messages[startIdx:endIdx] {
			start := time.Now()

			token := client.Publish(msg.Topic, byte(cfg.QoS), false, msg.Payload)
			token.Wait()

			if token.Error() != nil {
				// Фиксация ошибок
				met.IncError()
				logger.Log.Error().Msgf("Failed to publish message %q", msg)
			} else {
				// Фиксация размера нагрузки и топика
				met.IncSent(len(msg.Payload), len(msg.Topic))
				// Фиксация латенсии
				met.ObserveLatency(time.Since(start))
			}

			logger.Log.Debug().Msgf("Published message %q", msg)
		}
	}
}

func BurstMode(client mqtt.Client, cfg *config.Config, messages []generator.Message, met *metrics.Metrics) {
	for sent := 0; sent < len(messages); {
		batchStart := time.Now()

		for i := 0; i < cfg.BurstSize && sent < len(messages); i++ {
			msg := messages[sent]
			start := time.Now()

			token := client.Publish(msg.Topic, byte(cfg.QoS), false, msg.Payload)
			token.Wait()

			if token.Error() != nil {
				// Фиксация ошибок
				met.IncError()
				logger.Log.Error().Msgf("Failed to publish message %q", msg)
			} else {
				// Фиксация размера нагрузки и топика
				met.IncSent(len(msg.Payload), len(msg.Topic))
				// Фиксация латенсии
				met.ObserveLatency(time.Since(start))
				logger.Log.Debug().Msgf("Published message %q", msg)
			}
			sent++
		}
		// Расчет средней пропускной способности одной публикации в разрезе пачки
		elapsed := time.Since(batchStart).Seconds()
		if elapsed > 0.001 {
			met.SetThroughput(float64(cfg.BurstSize) / elapsed)
		}
		// Задержка
		logger.Log.Info().Msgf("Published burst %q", cfg.BurstSize)
		if sent < len(messages) {
			time.Sleep(time.Duration(cfg.BurstIntervalMs) * time.Millisecond)
		}
	}
}

func RampUpMode(client mqtt.Client, cfg *config.Config, messages []generator.Message, met *metrics.Metrics) {
	startRate := cfg.StartRate
	endRate := cfg.EndRate
	rampDuration := time.Duration(cfg.RampDurationSec) * time.Second
	totalSignals := cfg.TotalSignals

	startTime := time.Now()
	sent := 0

	for sent < totalSignals {
		elapsed := time.Since(startTime)
		if elapsed > rampDuration {
			elapsed = rampDuration
		}

		// Линейное увеличение rate
		currentRate := startRate + int(float64(endRate-startRate)*elapsed.Seconds()/rampDuration.Seconds())
		if currentRate < 1 {
			currentRate = 1
		}

		batchSize := currentRate
		if sent+batchSize > totalSignals {
			batchSize = totalSignals - sent
		}

		batchStart := time.Now()
		for i := 0; i < batchSize && sent < totalSignals; i++ {
			msg := messages[sent]
			start := time.Now()

			token := client.Publish(msg.Topic, byte(cfg.QoS), false, msg.Payload)
			token.Wait()

			if token.Error() != nil {
				met.IncError()
				logger.Log.Error().Msgf("Failed to publish message %q", msg)
			} else {
				met.IncSent(len(msg.Payload), len(msg.Topic))
				met.ObserveLatency(time.Since(start))
				logger.Log.Debug().Msgf("Published message %q", msg)
			}
			sent++
		}
		logger.Log.Info().Msgf("Published butch %d", int(batchSize))
		elapsedBatch := time.Since(batchStart).Seconds()
		if elapsedBatch > 0 {
			met.SetThroughput(float64(batchSize) / elapsedBatch)
		}

		// Поддержание заданного rate
		expectedTime := float64(batchSize) / float64(currentRate)
		actualTime := elapsedBatch
		if actualTime < expectedTime {
			time.Sleep(time.Duration((expectedTime-actualTime)*1000) * time.Millisecond)
		}
	}
}

func RandomMode(client mqtt.Client, cfg *config.Config, messages []generator.Message, met *metrics.Metrics) {
	minRate := cfg.MinRate
	maxRate := cfg.MaxRate
	totalSignals := cfg.TotalSignals
	sent := 0

	for sent < totalSignals {
		// Случайный rate между minRate и maxRate
		currentRate := minRate + rand.Intn(maxRate-minRate+1)
		if currentRate < 1 {
			currentRate = 1
		}

		batchSize := currentRate
		if sent+batchSize > totalSignals {
			batchSize = totalSignals - sent
		}

		batchStart := time.Now()
		for i := 0; i < batchSize && sent < totalSignals; i++ {
			msg := messages[sent]
			start := time.Now()

			token := client.Publish(msg.Topic, byte(cfg.QoS), false, msg.Payload)
			token.Wait()

			if token.Error() != nil {
				met.IncError()
				logger.Log.Error().Msgf("Failed to publish message %q", msg)
			} else {
				met.IncSent(len(msg.Payload), len(msg.Topic))
				met.ObserveLatency(time.Since(start))
				logger.Log.Debug().Msgf("Published message %q", msg)
			}
			sent++
		}
		logger.Log.Info().Msgf("Published butch %f", float64(batchSize))
		elapsed := time.Since(batchStart).Seconds()
		if elapsed > 0 {
			met.SetThroughput(float64(batchSize) / elapsed)
		}

		// Поддержание случайного rate
		expectedTime := float64(batchSize) / float64(currentRate)
		actualTime := elapsed
		if actualTime < expectedTime {
			time.Sleep(time.Duration((expectedTime-actualTime)*1000) * time.Millisecond)
		}
	}
}
