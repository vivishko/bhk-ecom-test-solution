package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository struct {
    db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
    return &Repository{db: db}
}

func (r *Repository) UpsertBannerStat(ctx context.Context, bannerID int, hour time.Time, counts map[string]int) error {
    countsJSON, err := json.Marshal(counts)
    if err != nil {
        return err
    }

    _, err = r.db.Exec(ctx, `
        INSERT INTO banner_stats (banner_id, hour_timestamp, counts)
        VALUES ($1, $2, $3::jsonb)
        ON CONFLICT (banner_id, hour_timestamp)
        DO UPDATE SET counts = EXCLUDED.counts;
    `, bannerID, hour, countsJSON)

    return err
}
