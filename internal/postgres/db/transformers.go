// Package db contains SQLC-generated database queries and type transformers.
package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// PgUUIDToUUID converts a pgtype.UUID to a google/uuid.UUID.
func PgUUIDToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}

	return uuid.UUID(id.Bytes)
}

// UUIDToPGUUID converts a google/uuid.UUID to a pgtype.UUID.
func UUIDToPGUUID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{}
	}

	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

// PgTimestamptzToTime converts a pgtype.Timestamptz to a time.Time.
func PgTimestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}

	return ts.Time
}

// PgTimestamptzToOptionalTime converts a pgtype.Timestamptz to a *time.Time, returning nil if invalid.
func PgTimestamptzToOptionalTime(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}

	t := ts.Time
	return &t
}

// TimeToPGTimestamptz converts a time.Time to a pgtype.Timestamptz.
func TimeToPGTimestamptz(ts time.Time) pgtype.Timestamptz {
	if ts.IsZero() {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{
		Time:  ts,
		Valid: true,
	}
}

// OptionalTimeToPGTimestamptz converts a *time.Time to a pgtype.Timestamptz, returning an invalid value if nil.
func OptionalTimeToPGTimestamptz(ts *time.Time) pgtype.Timestamptz {
	if ts == nil {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{
		Time:  *ts,
		Valid: true,
	}
}
