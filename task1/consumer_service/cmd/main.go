package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"consumer_service/internal/configs"
	"consumer_service/internal/handlers"
	"consumer_service/internal/repositories"

	_ "github.com/lib/pq"
)

func main() {
	logger := log.Default()

	dbpool, err := configs.NewPostgresDB()
    if err != nil {
        log.Fatalf("Failed to connect to database: %s", err)
    }
    defer dbpool.Close()

	repo := repositories.NewRepository(dbpool)

	kafkaCfg := configs.LoadKafkaConfig()
	consumerGroup, err := configs.NewConsumerGroup(kafkaCfg)
	if err != nil {
		logger.Fatalf("failed to create consumer group: %v", err)
	}
	defer consumerGroup.Close()

	aggregatedData := make(handlers.AggregatedData)
	mu := &sync.Mutex{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			handler := handlers.NewMessageHandler(mu, aggregatedData, logger)
			if err := consumerGroup.Consume(ctx, []string{kafkaCfg.KafkaTopic}, handler); err != nil {
				logger.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	minutesStr := os.Getenv("MINUTES_TICKER")
    if minutesStr == "" {
        minutesStr = "10" 
    }
    minutes, err := strconv.Atoi(minutesStr)
    if err != nil {
        logger.Printf("invalid MINUTES_TICKER value, using default 10")
        minutes = 10
    }

    ticker := time.NewTicker(time.Duration(minutes) * time.Minute)
    defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mu.Lock()
			err := flushDataToDB(ctx, repo, aggregatedData, logger)
			if err != nil {
				logger.Printf("error flushing data to DB: %v", err)
			}
			aggregatedData = make(handlers.AggregatedData)
			mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func flushDataToDB(ctx context.Context, repo *repositories.Repository, data handlers.AggregatedData, logger *log.Logger) error {
	for bannerID, hoursMap := range data {
		for hour, minutesMap := range hoursMap {
			counts := make(map[string]int)
			for m, count := range minutesMap {
				counts[strconv.Itoa(m)] = count
			}
			err := repo.UpsertBannerStat(ctx, bannerID, hour, counts)
			if err != nil {
				logger.Printf("failed to upsert: bannerID=%d hour=%s err=%v", bannerID, hour, err)
				return err
			}
		}
	}
	logger.Printf("Flushed aggregated data to DB at %s", time.Now())
	return nil
}