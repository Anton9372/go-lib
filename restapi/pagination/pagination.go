package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

var ErrInvalidCursor = errors.New("invalid pagination cursor")

type OffsetRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

func (p OffsetRequest) GetLimit() int {
	if p.Limit <= 0 {
		return defaultLimit
	}
	if p.Limit > maxLimit {
		return maxLimit
	}

	return p.Limit
}

type CursorRequest struct {
	Limit  int    `form:"limit"`
	Cursor string `form:"cursor"`
}

func (p CursorRequest) GetLimit() int {
	if p.Limit <= 0 {
		return defaultLimit
	}
	if p.Limit > maxLimit {
		return maxLimit
	}

	return p.Limit
}

type CursorData struct {
	CreatedAt time.Time `json:"c"`
	ID        uuid.UUID `json:"i"`
}

func DecodeCursor(encoded string) (*CursorData, error) {
	if encoded == "" {
		//nolint:nilnil // empty cursor is a valid case (first page). upstream code should check for CursorData == nil
		return nil, nil
	}

	decodedBytes, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: decode cursor: %v", ErrInvalidCursor, err) //nolint:errorlint // it's ok
	}

	var data CursorData
	if err = json.Unmarshal(decodedBytes, &data); err != nil {
		return nil, fmt.Errorf("%w: unmarshal cursor data: %v", ErrInvalidCursor, err) //nolint:errorlint // it's ok
	}

	return &data, nil
}

func EncodeCursor(createdAt time.Time, id uuid.UUID) (string, error) {
	data := CursorData{
		CreatedAt: createdAt,
		ID:        id,
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal cursor data: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

type ResponsePage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	Total      int64  `json:"total,omitempty"`
}
