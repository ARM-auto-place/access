package custom_handlers

import (
	"errors"

	customErrors "github.com/ARM-auto-place/access/pkg/models/errors"
)

// ErrorHandler prepares the public error for the generated JSON-RPC response.
// The generated service logger receives the original error before this handler.
func ErrorHandler(err error) error {
	if err == nil {
		return nil
	}

	var customErr *customErrors.Error
	if errors.As(err, &customErr) {
		return customErrors.New(customErr.GetPublicMessage(), customErr.GetStatusCode()).WithCode(customErr.Code())
	}

	return customErrors.InternalServerError()
}
