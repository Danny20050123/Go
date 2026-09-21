package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var database *sql.DB

func openDatabase(ctx context.Context, databaseURL string) (*sql.DB, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	pingContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingContext); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

// saveHighScore inserts a new player score or replaces an existing score only
// when the new score is higher.
func saveHighScore(ctx context.Context, playerID string, score int) error {
	if database == nil {
		return errors.New("database is not connected")
	}

	const query = `
		INSERT INTO high_scores (player_id, high_score)
		VALUES ($1, $2)
		ON CONFLICT (player_id) DO UPDATE
		SET
			high_score = EXCLUDED.high_score,
			achieved_at = CURRENT_TIMESTAMP
		WHERE EXCLUDED.high_score > high_scores.high_score
	`

	if _, err := database.ExecContext(ctx, query, playerID, score); err != nil {
		return fmt.Errorf("save high score: %w", err)
	}

	return nil
}
