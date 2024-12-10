package handlers

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type ClickMessage struct {
	BannerID  int       `json:"banner_id"`
	Timestamp time.Time `json:"timestamp"`
}

type AggregatedData map[int]map[time.Time]map[int]int

type MessageHandler struct {
	mu             *sync.Mutex
	aggregatedData AggregatedData
	logger         *log.Logger
}

func NewMessageHandler(mu *sync.Mutex, aggregatedData AggregatedData, logger *log.Logger) *MessageHandler {
    return &MessageHandler{
        mu:  mu,
        aggregatedData: aggregatedData,
		logger: logger,
    }
}

func (h *MessageHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *MessageHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *MessageHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var cm ClickMessage
		if err := json.Unmarshal(msg.Value, &cm); err != nil {
			h.logger.Printf("failed to unmarshal message: %v", err)
			continue
		}
		h.mu.Lock()
		hour := cm.Timestamp.Truncate(time.Hour)
		minute := cm.Timestamp.Minute()

		if h.aggregatedData[cm.BannerID] == nil {
			h.aggregatedData[cm.BannerID] = make(map[time.Time]map[int]int)
		}
		if h.aggregatedData[cm.BannerID][hour] == nil {
			h.aggregatedData[cm.BannerID][hour] = make(map[int]int)
		}
		h.aggregatedData[cm.BannerID][hour][minute]++
		h.mu.Unlock()

		session.MarkMessage(msg, "")
	}
	return nil
}