package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"producer_service/internal/configs"
	"producer_service/internal/repositories"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ClickMessage struct {
	BannerID  int       `json:"banner_id"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of request durations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	fmt.Println("Initializing Prometheus metrics")
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
}

func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := c.Writer.Status()

		requestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), strconv.Itoa(statusCode)).Inc()
		requestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

func main() {
	logger := log.Default()

	dbpool, err := configs.NewPostgresDB()
    if err != nil {
        log.Fatalf("Failed to connect to database: %s", err)
    }
    defer dbpool.Close()

	repo := repositories.NewRepository(dbpool)

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{kafkaBroker}, config)
	if err != nil {
		logger.Fatalf("failed to create kafka producer: %v", err)
	}
	defer producer.Close()

	r := gin.Default()
	r.Use(prometheusMiddleware()) 

	r.GET("/metrics", gin.WrapH(promhttp.Handler())) 

	r.POST("/counter/:bannerID", func(c *gin.Context) {
		bidStr := c.Param("bannerID")
		bid, err := strconv.Atoi(bidStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bannerID"})
			return
		}

		msg := ClickMessage{
			BannerID:  bid,
			Timestamp: time.Now(),
		}
		data, err := json.Marshal(msg)
		if err != nil {
			logger.Printf("failed to marshal message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: kafkaTopic,
			Value: sarama.ByteEncoder(data),
		})
		if err != nil {
			logger.Printf("failed to send message to kafka: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not send message"})
			return
		}
		logger.Printf("message sent to kafka partition=%d, offset=%d", partition, offset)
		c.JSON(http.StatusOK, gin.H{"status": "message sent"})
	})

	r.GET("/stats/:bannerID", func(c *gin.Context) {
		bidStr := c.Param("bannerID")
		bid, err := strconv.Atoi(bidStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bannerID"})
			return
		}

		var req struct {
			TsFrom time.Time `json:"tsFrom"`
			TsTo   time.Time `json:"tsTo"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		stat, err := repo.GetStats(context.Background(), bid, req.TsFrom, req.TsTo)
		if err != nil {
			if err.Error() == "no data found for this timestamp" {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				logger.Printf("failed to get stats: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve stats"})
			}
			return
		}

		c.JSON(http.StatusOK, stat)
	})

	r.Run(":8080")
}