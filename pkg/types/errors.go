package types

import "fmt"

type ErrorType int

const (
	ErrInvalidInput ErrorType = iota
	ErrInvalidFormat
	ErrMessageTooLong
	ErrProcessingFailed
)

type Error struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewError(errType ErrorType, message string, err error) error {
	return &Error{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}
