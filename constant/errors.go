package constant

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorizedAccess = errors.New("You are not permitted to perform this action")
	ErrPathNotFound       = errors.New("file or Directory source does not exist")
	ErrPathFormatNotFound = errors.New("Error: path %s does not exist")
	ErrTokenNotFound      = errors.New("failed to get token from context")
	ErrHostNotFound       = errors.New("failed to get host from context")
	ErrClientsNotFound    = errors.New("failed to get client list from context")
)

func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}
