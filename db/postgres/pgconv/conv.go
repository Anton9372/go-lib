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

func Int64(i pgtype.Int8) (int64, error) {
	if !i.Valid {
		return 0, errors.New("int8 is null") //nolint:err113 // errors.New is perfect here
	}

	return i.Int64, nil
}

func Int64Ptr(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}

	return &i.Int64
}

func PgInt8(i int64) pgtype.Int8 {
	return pgtype.Int8{
		Int64: i,
		Valid: true,
	}
}

func PgInt8Ptr(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}

	return pgtype.Int8{
		Int64: *i,
		Valid: true,
	}
}
