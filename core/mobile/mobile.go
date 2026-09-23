package mobile

import (
	"errors"

	"obxod/internal/clienthello"
	"obxod/internal/plan"
	"obxod/internal/rules"
	"obxod/internal/sweep"
)

var (
	ErrPanicked = errors.New("mobile: recovered from a panic in core")
	ErrNoRule   = errors.New("mobile: no rule for the host")
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

func PlanForStrategy(hello []byte, strategy string) (*Plan, error) {
	return guard(func() (*Plan, error) {
		// rules.Parse wants host=ways; the host is unused when planning a bare strategy.
		rule, err := rules.Parse("x=" + strategy)
		if err != nil {
			return nil, err
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

func Strategies() *Ways {
	return &Ways{names: sweep.Strategies()}
}

type Ways struct {
	names []string
}

func (w *Ways) Count() int {
	return len(w.names)
}

func (w *Ways) At(index int) string {
	if index < 0 || index >= len(w.names) {
		return ""
	}

	return w.names[index]
}

type Sweep struct {
	best  string
	bytes int
}

func NewSweep() *Sweep {
	return &Sweep{}
}

func (s *Sweep) Record(strategy string, bytes int) {
	if bytes > s.bytes {
		s.best, s.bytes = strategy, bytes
	}
}

func (s *Sweep) Best() string {
	return s.best
}

func (s *Sweep) Found() bool {
	return s.bytes > 0
}
