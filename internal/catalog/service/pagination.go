package service

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/artlink52/marketplace/internal/catalog/domain"
	"github.com/google/uuid"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Токен страницы непрозрачен для клиента: base64("<created_at>|<id>").
// Клиент просто возвращает его в следующем запросе.
func encodeCursor(c domain.ProductCursor) string {
	raw := c.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + c.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(token string) (domain.ProductCursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return domain.ProductCursor{}, err
	}

	createdAtStr, idStr, ok := strings.Cut(string(raw), "|")
	if !ok {
		return domain.ProductCursor{}, errors.New("malformed cursor")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
		return domain.ProductCursor{}, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return domain.ProductCursor{}, err
	}

	return domain.ProductCursor{CreatedAt: createdAt, ID: id}, nil
}

func normalizePageSize(size int32) int32 {
	switch {
	case size <= 0:
		return defaultPageSize
	case size > maxPageSize:
		return maxPageSize
	default:
		return size
	}
}
