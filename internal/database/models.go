// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

import (
	"database/sql"
	"time"
)

type Feed struct {
	ID            int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Name          string
	Url           string
	UserID        int64
	LastFetchedAt sql.NullTime
}

type FeedFollow struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    int64
	FeedID    int64
}

type User struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
}
