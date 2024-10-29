package storage

import "errors"

var (
	ErrURLExists = errors.New("url already exists")

	ErrURLNotFound = errors.New("url not found")
)
