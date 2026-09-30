package models

import "errors"

var (
	ErrNotFound     = errors.New("models: resource could not be found")
	ErrUnauthorized = errors.New("models: unauthorized to access resource")
)
