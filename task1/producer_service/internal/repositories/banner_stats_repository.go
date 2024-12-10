package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type BannerStatResponse struct {
    BannerID int    `json:"bannerID"`
    TsFrom   string `json:"tsFrom"`
    TsTo     string `json:"tsTo"`
    Counts   int    `json:"Counts"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetStats(ctx context.Context, bannerID int, tsFrom, tsTo time.Time) (*BannerStatResponse, error) {
    if tsFrom.After(tsTo) {
        return nil, fmt.Errorf("tsFrom should be before tsTo")
    }

    startHour := tsFrom.Truncate(time.Hour)
    endHour := tsTo.Truncate(time.Hour)
    if tsTo.Minute() > 0 || tsTo.Second() > 0 || tsTo.Nanosecond() > 0 {
        endHour = endHour.Add(time.Hour)
    }

    rows, err := r.db.Query(ctx, `
        SELECT hour_timestamp, counts
        FROM banner_stats
        WHERE banner_id = $1
          AND hour_timestamp >= $2
          AND hour_timestamp < $3
    `, bannerID, startHour, endHour)
    if err != nil {
        return nil, fmt.Errorf("failed to query database: %w", err)
    }
    defer rows.Close()

    countsByHour := make(map[time.Time]map[string]int)
    for rows.Next() {
        var hourTimestamp time.Time
        var countsJSON []byte
        if err := rows.Scan(&hourTimestamp, &countsJSON); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }

        var counts map[string]int
        if err := json.Unmarshal(countsJSON, &counts); err != nil {
            return nil, fmt.Errorf("failed to unmarshal JSON counts: %w", err)
        }

        countsByHour[hourTimestamp] = counts
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("failed to iterate by rows: %w", err)
    }

    if len(countsByHour) == 0 {
        return nil, fmt.Errorf("no data found for this timestamp")
    }

    totalCounts := 0
    for hour, counts := range countsByHour {
        currentHourStart := hour
        currentHourEnd := hour.Add(time.Hour)

        intervalStart := tsFrom
        if currentHourStart.After(tsFrom) {
            intervalStart = currentHourStart
        }
        intervalEnd := tsTo
        if currentHourEnd.Before(tsTo) {
            intervalEnd = currentHourEnd
        }

        startMinute := intervalStart.Minute()
        endMinute := intervalEnd.Minute()
        if intervalEnd.Equal(currentHourEnd) {
            endMinute = 60
        }

        for m := startMinute; m < endMinute; m++ {
            key := fmt.Sprintf("%d", m)
            if count, exists := counts[key]; exists {
                totalCounts += count
            }
        }
    }

    response := &BannerStatResponse{
        BannerID: bannerID,
        TsFrom:   tsFrom.Format(time.RFC3339),
        TsTo:     tsTo.Format(time.RFC3339),
        Counts:   totalCounts,
    }

    return response, nil
}