package plan

import (
	"errors"
	"fmt"
	"strings"

	"obxod/internal/clienthello"
	"obxod/internal/rules"
)

// The decoy must die within a few hops, before the server but past the inspector.
const decoyTTL = 8

var (
	ErrNoSocketWay = errors.New("plan: the rule names no way a socket can apply")
	ErrNotOnSocket = errors.New("plan: the rule needs a way the socket path cannot do")
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
	if way := notOnSocket(rule); way != "" {
		return Plan{}, fmt.Errorf("%w: %s", ErrNotOnSocket, way)
	}

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

		times := rule.Repeats
		if times == 0 {
			times = 1
		}
		for i := 0; i < times; i++ {
			segments = append(segments, Segment{Bytes: doomed, TTL: ttl})
		}
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

func notOnSocket(rule rules.Rule) string {
	switch {
	case rule.BadSeq != 0:
		return "badseq"
	case rule.BadAck != 0:
		return "badack"
	case rule.BadSum:
		return "badsum"
	case rule.Signed:
		return "md5sig"
	case rule.Recorded:
		return "fake"
	case rule.Overlap != 0:
		return "overlap"
	case rule.Stale != 0:
		return "ts"
	case rule.HostFake != "":
		return "hostfake"
	case rule.FakeUDP != 0:
		return "fakeudp"
	}

	return ""
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
