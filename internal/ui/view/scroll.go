package view

type Scroll struct {
	size  int
	start int
}

func (s *Scroll) Resize(size int) {
	s.size = max(size, 0)
}

func (s *Scroll) Top() {
	s.start = 0
}

func (s *Scroll) Follow(cursor, total int) {
	start, end := s.Range(total)

	switch {
	case cursor < start:
		s.start = cursor
	case cursor >= end:
		s.start = cursor - s.size + 1
	}

	s.clamp(total)
}

func (s *Scroll) Range(total int) (int, int) {
	s.clamp(total)

	if s.windowed(total) {
		return s.start, s.start + s.size
	}

	return 0, max(total, 0)
}

func (s *Scroll) Above() int {
	return s.start
}

func (s *Scroll) Below(total int) int {
	_, end := s.Range(total)

	return max(total-end, 0)
}

func (s *Scroll) windowed(total int) bool {
	return s.size > 0 && total > s.size
}

func (s *Scroll) clamp(total int) {
	if !s.windowed(total) {
		s.start = 0
		return
	}

	s.start = min(max(s.start, 0), total-s.size)
}
