package app

import (
	"strconv"
)

const (
	DefaultPageLimit = 50
	MaxPageLimit     = 100
)

type PageParams struct {
	Limit  int
	Offset int
}

type PageResult[T any] struct {
	Items      []T
	TotalCount int
}

func ParsePageParams(limitRaw string, offsetRaw string) PageParams {
	limit := DefaultPageLimit
	offset := 0

	if limitRaw != "" {
		if parsed, err := strconv.Atoi(limitRaw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	if offsetRaw != "" {
		if parsed, err := strconv.Atoi(offsetRaw); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return PageParams{Limit: limit, Offset: offset}
}
