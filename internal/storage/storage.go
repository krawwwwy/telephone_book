package storage

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrSchemaNotExist    = errors.New("schema not exists")
)
