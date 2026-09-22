package plan

import "obxod/internal/clienthello"

type Segment struct {
	Bytes []byte
	TTL   int // 0 leaves the socket default; a positive value is set for this one segment
}

type Plan struct {
	Segments []Segment
}

func Cut(hello []byte) (Plan, error) {
	parsed, err := clienthello.Parse(hello)
	if err != nil {
		return Plan{}, err
	}

	found, err := parsed.ServerName()
	if err != nil {
		return Plan{}, err
	}

	at := splitWithin(found, len(hello))

	return Plan{Segments: []Segment{
		{Bytes: hello[:at]},
		{Bytes: hello[at:]},
	}}, nil
}

func splitWithin(sni clienthello.ServerName, size int) int {
	at := sni.Offset + len(sni.Host)/2
	if at <= 0 {
		at = 1
	}
	if at >= size {
		at = size - 1
	}
	return at
}
