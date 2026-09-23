package sweep

import "strconv"

// The socket-doable rows of the desktop sweep: split modes, and a decoy that dies
// at a range of hop counts, with and without a following split.
func Strategies() []string {
	out := []string{"cut:name", "cut:start", "cut:after"}

	for _, hops := range []int{1, 2, 3, 4, 6, 8} {
		ttl := "decoy:auto,ttl:" + strconv.Itoa(hops)
		out = append(out, ttl+",cut:name", ttl)
	}

	return out
}
