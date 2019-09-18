package client

import "fmt"

type Error struct {
	err     error
	msg, op string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.op, e.msg)
}

func (e *Error) Err(err error) *Error {
	e.err = err
	return e
}

func (e *Error) Msg(msg string) *Error {
	e.msg = msg
	return e
}

func (e *Error) Unwrap() error {
	return e.err
}
