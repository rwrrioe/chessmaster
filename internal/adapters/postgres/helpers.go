package postgres

import "errors"

var errNotFound = errors.New("not found")

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
