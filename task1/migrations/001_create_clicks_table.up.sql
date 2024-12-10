CREATE TABLE IF NOT EXISTS banner_stats (
  id SERIAL PRIMARY KEY,
  banner_id INT NOT NULL,
  hour_timestamp TIMESTAMP NOT NULL,
  counts JSONB NOT NULL,
  UNIQUE (banner_id, hour_timestamp)
);