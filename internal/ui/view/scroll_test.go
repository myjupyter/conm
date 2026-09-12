package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScroll(t *testing.T) {
	type follow struct {
		cursor int
		total  int
	}

	tests := map[string]struct {
		size    int
		resize  int
		follows []follow
		top     bool
		total   int
		start   int
		end     int
		above   int
		below   int
	}{
		"fits in window": {
			size:    5,
			follows: []follow{{cursor: 2, total: 3}},
			total:   3, start: 0, end: 3, above: 0, below: 0,
		},
		"no size shows everything": {
			follows: []follow{{cursor: 7, total: 9}},
			total:   9, start: 0, end: 9, above: 0, below: 0,
		},
		"cursor inside leaves window alone": {
			size:    3,
			follows: []follow{{cursor: 5, total: 10}, {cursor: 4, total: 10}},
			total:   10, start: 3, end: 6, above: 3, below: 4,
		},
		"drags down by one": {
			size:    3,
			follows: []follow{{cursor: 2, total: 10}, {cursor: 3, total: 10}},
			total:   10, start: 1, end: 4, above: 1, below: 6,
		},
		"drags up by one": {
			size:    3,
			follows: []follow{{cursor: 9, total: 10}, {cursor: 6, total: 10}},
			total:   10, start: 6, end: 9, above: 6, below: 1,
		},
		"jump to bottom": {
			size:    4,
			follows: []follow{{cursor: 9, total: 10}},
			total:   10, start: 6, end: 10, above: 6, below: 0,
		},
		"jump to top": {
			size:    4,
			follows: []follow{{cursor: 9, total: 10}, {cursor: 0, total: 10}},
			total:   10, start: 0, end: 4, above: 0, below: 6,
		},
		"top resets": {
			size:    3,
			follows: []follow{{cursor: 9, total: 10}},
			top:     true,
			total:   10, start: 0, end: 3, above: 0, below: 7,
		},
		"shrinking total clamps": {
			size:    3,
			follows: []follow{{cursor: 9, total: 10}},
			total:   5, start: 2, end: 5, above: 2, below: 0,
		},
		"total below start clamps": {
			size:    3,
			follows: []follow{{cursor: 9, total: 10}},
			total:   2, start: 0, end: 2, above: 0, below: 0,
		},
		"growing size clamps": {
			size:    3,
			follows: []follow{{cursor: 9, total: 10}},
			resize:  8,
			total:   10, start: 2, end: 10, above: 2, below: 0,
		},
		"empty total": {
			size:    3,
			follows: []follow{{cursor: 0, total: 0}},
			total:   0, start: 0, end: 0, above: 0, below: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var s Scroll

			s.Resize(tt.size)

			for _, f := range tt.follows {
				s.Follow(f.cursor, f.total)
			}

			if tt.resize > 0 {
				s.Resize(tt.resize)
			}

			if tt.top {
				s.Top()
			}

			start, end := s.Range(tt.total)

			is := assert.New(t)
			is.Equal(tt.start, start)
			is.Equal(tt.end, end)
			is.Equal(tt.above, s.Above())
			is.Equal(tt.below, s.Below(tt.total))
		})
	}
}
