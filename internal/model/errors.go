package model

import "errors"

var (
	ErrInvalidGroupBy         = errors.New("invalid grouping parameter specified")
	ErrInvalidOrderBy         = errors.New("invalid ordering parameter specified")
	ErrOperationIDNotFound    = errors.New("specified operation ID not found")
	ErrUnknownActorOrCategory = errors.New("invalid actor or category")
)
