package pgconv

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func UUID(id pgtype.UUID) (uuid.UUID, error) {
	if !id.Valid {
		return uuid.Nil, errors.New("uuid is null") //nolint:err113 // errors.New is perfect here
	}

	return id.Bytes, nil
}

func UUIDPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}

	u := uuid.UUID(id.Bytes)

	return &u
}

func PgUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

func PgUUIDPtr(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}

	return pgtype.UUID{
		Bytes: *u,
		Valid: true,
	}
}

func Text(t pgtype.Text) (string, error) {
	if !t.Valid {
		return "", errors.New("text is null") //nolint:err113 // errors.New is perfect here
	}

	return t.String, nil
}

func TextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}

	return &t.String
}

func PgText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

func PgTextPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}

	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

func Time(t pgtype.Timestamptz) (time.Time, error) {
	if !t.Valid {
		return time.Time{}, errors.New("timestamptz is null") //nolint:err113 // errors.New is perfect here
	}

	return t.Time, nil
}

func TimePtr(t pgtype.Timestamptz) *time.Time {
	var ptr *time.Time

	if t.Valid {
		ptr = &t.Time
	}

	return ptr
}

func PgTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

func PgTimePtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}

	return pgtype.Timestamptz{
		Time:  *t,
		Valid: true,
	}
}
