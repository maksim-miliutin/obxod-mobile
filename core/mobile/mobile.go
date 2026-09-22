package mobile

import (
	"errors"

	"obxod/internal/clienthello"
)

var ErrPanicked = errors.New("mobile: recovered from a panic in core")

func ServerName(hello []byte) (name string, err error) {
	// A panic crossing the gomobile boundary exits the app, so it becomes an error here.
	defer func() {
		if recover() != nil {
			name, err = "", ErrPanicked
		}
	}()

	parsed, err := clienthello.Parse(hello)
	if err != nil {
		return "", err
	}

	found, err := parsed.ServerName()
	if err != nil {
		return "", err
	}

	return found.Host, nil
}
