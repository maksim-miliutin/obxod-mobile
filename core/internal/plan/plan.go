package plan

import (
	"errors"
	"strings"

	"obxod/internal/clienthello"
	"obxod/internal/rules"
)

// The decoy must die within a few hops, before the server but past the inspector.
const decoyTTL = 8

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
	if rule.Decoy == "" && rule.Cut == "" {
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

	var segments []Segment

	if rule.Decoy != "" {
		doomed, err := decoy(parsed, found.Host, rule.Decoy)
		if err != nil {
			return Plan{}, err
		}

		ttl := int(rule.TTL)
		if ttl == 0 {
			ttl = decoyTTL
		}

		segments = append(segments, Segment{Bytes: doomed, TTL: ttl})
	}

	real, err := realSegments(hello, found, rule.Cut, rule.Disorder)
	if err != nil {
		return Plan{}, err
	}

	return Plan{Segments: append(segments, real...)}, nil
}

func decoy(parsed clienthello.Hello, host, spec string) ([]byte, error) {
	name := spec
	if spec == "auto" {
		name = decoyName(host)
	}

	return parsed.Renamed(name)
}

// A harmless name exactly as long as the real one, so the lengths inside the hello stay true.
func decoyName(host string) string {
	const base = "google.com"

	if len(host) < len(base)+2 {
		return strings.Repeat("a", len(host)-4) + ".com"
	}

	return strings.Repeat("x", len(host)-len(base)-1) + "." + base
}

func realSegments(hello []byte, sni clienthello.ServerName, cut string, disorder bool) ([]Segment, error) {
	if cut == "" {
		return []Segment{{Bytes: hello}}, nil
	}

	at, err := splitPoint(sni, cut, len(hello))
	if err != nil {
		return nil, err
	}

	first := Segment{Bytes: hello[:at]}
	if disorder {
		// A doomed first part: the inspector sees it, the kernel resends it once the next
		// segment's normal TTL has restored the default.
		first.TTL = 1
	}

	return []Segment{first, {Bytes: hello[at:]}}, nil
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
