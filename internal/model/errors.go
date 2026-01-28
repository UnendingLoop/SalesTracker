package model

import "errors"

var (
	ErrCommon500              = errors.New("something went wrong, try again later")
	ErrInvalidGroupBy         = errors.New("invalid grouping parameter specified")
	ErrInvalidOrderBy         = errors.New("invalid ordering parameter specified")
	ErrOperationIDNotFound    = errors.New("specified operation ID not found")
	ErrUnknownActorOrCategory = errors.New("invalid actor or category provided")
	ErrInvalidAmount          = errors.New("invalid amount provided")
	ErrInvalidActor           = errors.New("invalid actor provided")
	ErrInvalidCategory        = errors.New("invalid category provided")
	ErrInvalidOpType          = errors.New("invalid operation type provided")
	ErrInvalidOpTime          = errors.New("invalid operation time provided")
	ErrInvalidID              = errors.New("invalid operation ID provided")
	ErrInvalidAscDesc         = errors.New("invalid ASC/DESC provided")
	ErrInvalidStartEndTime    = errors.New("invalid start/end time provided: start cannot be later than end")
	ErrInvalidPage            = errors.New("invalid page value provided: value must be > 0")
	ErrInvalidLimit           = errors.New("invalid limit value provided: value must be > 0 and < 1000")
)
