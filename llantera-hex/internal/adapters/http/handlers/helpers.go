package handlers

import (
	"strconv"
	"strings"
)

func extractResourceID(path, prefix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.Trim(trimmed, "/")
	return trimmed
}

func parseLimit(raw string) int {
	if raw == "" {
		return 20
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 20
	}
	if value > 10000 {
		value = 10000
	}
	return value
}

func parseOffset(raw string) int {
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func parsePositiveInt(raw string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}
