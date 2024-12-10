package configs

import (
	"log"
	"os"

	"github.com/IBM/sarama"
)

type KafkaConfig struct {
	KafkaBroker string
	KafkaTopic  string
}

func LoadKafkaConfig() *KafkaConfig {
	kafkaBroker := os.Getenv("KAFKA_BROKER") 
	if kafkaBroker == "" {
		log.Fatal("KAFKA_BROKER not set")
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC") 
	if kafkaTopic == "" {
		log.Fatal("KAFKA_TOPIC not set")
	}

	return &KafkaConfig{
		KafkaBroker: kafkaBroker,
		KafkaTopic:  kafkaTopic,
	}
}

func NewConsumerGroup(kCfg *KafkaConfig) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup([]string{kCfg.KafkaBroker}, "my_consumer_group", config)
	if err != nil {
		return nil, err
	}
	return consumerGroup, nil
}