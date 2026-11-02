package errors

import "errors"

var (
	ErrRoleNotFound   = errors.New("role not assigned")
	ErrUserNotFound   = errors.New("user not found")
	ErrRefreshAborted = errors.New("cache refresh aborted")
)
