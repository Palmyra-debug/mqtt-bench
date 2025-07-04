package generator

import (
	"fmt"
	"math/rand"
	"mqtt-bench/config"
	"mqtt-bench/logger"
	"strings"
)

type Message struct {
	Topic   string
	Payload string
}

func GenMessages(cfg *config.Config) []Message {
	messages := make([]Message, 0, cfg.TotalSignals)

	for i := 0; i < cfg.TotalSignals; i++ {
		topic := generateTopic(cfg.TopicPattern)
		payload := generatePayload()

		msg := Message{
			Topic:   topic,
			Payload: payload,
		}

		messages = append(messages, msg)
	}
	logger.Log.Info().Msgf("All messages already ready for publish %d", len(messages))
	return messages
}

func generateTopic(TopicPattern string) string {
	deviceID := rand.Intn(1000) + 1
	controlID := rand.Intn(10) + 1
	topic := strings.ReplaceAll(TopicPattern, "{device_id}", fmt.Sprintf("%d", deviceID))
	topic = strings.ReplaceAll(topic, "{control_id}", fmt.Sprintf("%d", controlID))
	return topic
}

func generatePayload() string {
	value := rand.Intn(101)
	return fmt.Sprintf("%d", value)
}
