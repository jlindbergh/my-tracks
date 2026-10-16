// Package store holds the database access for users and activities.
//
// Every activity query takes the owning user's id and filters on it, so a
// caller can never read or change another user's activities.
package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jlindbergh/my-tracks/backend/internal/gpxfile"
)

// Store wraps a PostgreSQL connection pool.
type Store struct {
	pool *pgxpool.Pool
}

// New connects to the database at databaseURL.
func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close closes the connection pool.
func (s *Store) Close() {
	s.pool.Close()
}

// User is an account in the application.
type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

// EnsureUserByEmail returns the user with the given email, creating it if needed.
func (s *Store) EnsureUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email) VALUES ($1)
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		RETURNING id, email, display_name`, email,
	).Scan(&u.ID, &u.Email, &u.DisplayName)
	if err != nil {
		return User{}, fmt.Errorf("ensure user %q: %w", email, err)
	}
	return u, nil
}

// NewActivity is an activity about to be stored.
type NewActivity struct {
	Activity       gpxfile.Activity
	SourceFilename string
	ContentSHA256  string
}

// InsertActivity stores an activity for userID. created is false, and id
// empty, when the user has already imported a file with the same content.
func (s *Store) InsertActivity(ctx context.Context, userID string, na NewActivity) (id string, created bool, err error) {
	a := na.Activity
	err = s.pool.QueryRow(ctx, `
		INSERT INTO activities
			(user_id, name, sport, started_at, duration_s, source, source_filename, content_sha256, geom)
		VALUES ($1, $2, $3, $4, $5, 'gpx', $6, $7, ST_GeomFromText($8, 4326))
		ON CONFLICT (user_id, content_sha256) DO NOTHING
		RETURNING id`,
		userID, a.Name, a.Sport, a.StartedAt, a.DurationS, na.SourceFilename, na.ContentSHA256, a.WKT(),
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("insert activity: %w", err)
	}
	return id, true, nil
}

// ActivitySummary is an activity without its geometry.
type ActivitySummary struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Sport     string     `json:"sport"`
	StartedAt *time.Time `json:"startedAt"`
	DurationS *int       `json:"durationS"`
	DistanceM float64    `json:"distanceM"`
	// BBox is [west, south, east, north] in WGS 84.
	BBox [4]float64 `json:"bbox"`
}

// ListActivities returns userID's activities, newest first.
func (s *Store) ListActivities(ctx context.Context, userID string) ([]ActivitySummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, sport, started_at, duration_s, distance_m,
		       ST_XMin(geom), ST_YMin(geom), ST_XMax(geom), ST_YMax(geom)
		  FROM activities
		 WHERE user_id = $1
		 ORDER BY started_at DESC NULLS LAST, imported_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list activities: %w", err)
	}
	defer rows.Close()

	list := []ActivitySummary{}
	for rows.Next() {
		var a ActivitySummary
		if err := rows.Scan(&a.ID, &a.Name, &a.Sport, &a.StartedAt, &a.DurationS, &a.DistanceM,
			&a.BBox[0], &a.BBox[1], &a.BBox[2], &a.BBox[3]); err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list activities: %w", err)
	}
	return list, nil
}
