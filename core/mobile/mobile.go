package mobile

import (
	"errors"

	"obxod/internal/clienthello"
)

var ErrPanicked = errors.New("mobile: recovered from a panic in core")

// A panic crossing the gomobile boundary exits the app, so guard turns it into an error.
func guard[T any](call func() (T, error)) (result T, err error) {
	defer func() {
		if recover() != nil {
			var zero T
			result, err = zero, ErrPanicked
		}
	}()

	return call()
}

func ServerName(hello []byte) (string, error) {
	return guard(func() (string, error) {
		parsed, err := clienthello.Parse(hello)
		if err != nil {
			return "", err
		}

		found, err := parsed.ServerName()
		if err != nil {
			return "", err
		}

		return found.Host, nil
	})
}
