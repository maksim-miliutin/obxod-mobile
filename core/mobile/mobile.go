package mobile

import (
	"errors"

	"obxod/internal/clienthello"
	"obxod/internal/plan"
	"obxod/internal/rules"
)

var (
	ErrPanicked   = errors.New("mobile: recovered from a panic in core")
	ErrUnknownWay = errors.New("mobile: unknown way")
	ErrNoRule     = errors.New("mobile: no rule for the host")
)

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

func PlanForWay(hello []byte, way string) (*Plan, error) {
	return guard(func() (*Plan, error) {
		switch way {
		case "cut":
			computed, err := plan.Cut(hello)
			if err != nil {
				return nil, err
			}

			return newPlan(computed.Segments), nil
		default:
			return nil, ErrUnknownWay
		}
	})
}

func PlanForHost(hello []byte, host string, rulesText string) (*Plan, error) {
	return guard(func() (*Plan, error) {
		set, err := rules.Several(rulesText)
		if err != nil {
			return nil, err
		}

		rule, ok := set.For(host)
		if !ok {
			return nil, ErrNoRule
		}

		computed, err := plan.FromRule(hello, rule)
		if err != nil {
			return nil, err
		}

		return newPlan(computed.Segments), nil
	})
}

type Plan struct {
	segments []plan.Segment
}

// Byte slices cross gomobile by reference; the copy keeps a reused caller buffer from rewriting the plan.
func newPlan(segments []plan.Segment) *Plan {
	own := make([]plan.Segment, len(segments))
	for i, s := range segments {
		own[i] = plan.Segment{Bytes: append([]byte(nil), s.Bytes...), TTL: s.TTL}
	}

	return &Plan{segments: own}
}

func (p *Plan) Count() int {
	return len(p.segments)
}

func (p *Plan) Bytes(index int) []byte {
	if index < 0 || index >= len(p.segments) {
		return nil
	}

	return p.segments[index].Bytes
}

func (p *Plan) TTL(index int) int {
	if index < 0 || index >= len(p.segments) {
		return 0
	}

	return p.segments[index].TTL
}
