package plan

import (
	"errors"

	"obxod/internal/clienthello"
	"obxod/internal/rules"
)

var (
	ErrNoSocketWay = errors.New("plan: the rule names no way a socket can apply")
	ErrBadCut      = errors.New("plan: cut wants name, after or start")
)

type Segment struct {
	Bytes []byte
	TTL   int // 0 leaves the socket default; a positive value is set for this one segment
}

type Plan struct {
	Segments []Segment
}

func Cut(hello []byte) (Plan, error) {
	return FromRule(hello, rules.Rule{Cut: "name"})
}

func FromRule(hello []byte, rule rules.Rule) (Plan, error) {
	if rule.Cut == "" {
		return Plan{}, ErrNoSocketWay
	}

	parsed, err := clienthello.Parse(hello)
	if err != nil {
		return Plan{}, err
	}

	found, err := parsed.ServerName()
	if err != nil {
		return Plan{}, err
	}

	at, err := splitPoint(found, rule.Cut, len(hello))
	if err != nil {
		return Plan{}, err
	}

	return Plan{Segments: []Segment{
		{Bytes: hello[:at]},
		{Bytes: hello[at:]},
	}}, nil
}

func splitPoint(sni clienthello.ServerName, mode string, size int) (int, error) {
	at := 0
	switch mode {
	case "name":
		at = sni.Offset + len(sni.Host)/2
	case "after":
		at = sni.Offset + len(sni.Host)
	case "start":
		at = 2
	default:
		return 0, ErrBadCut
	}

	if at <= 0 {
		at = 1
	}
	if at >= size {
		at = size - 1
	}

	return at, nil
}
