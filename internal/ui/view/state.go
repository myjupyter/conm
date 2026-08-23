package view

import "slices"

type tableState[K comparable] struct {
	idx index[K]

	row        int
	confirming bool
	help       bool
	searching  bool
}

func (s *tableState[K]) register(kind K) bool {
	return s.idx.register(kind)
}

func (s *tableState[K]) reset(entries []entry[K]) {
	s.idx.reset(entries)
	s.Sync()
}

func (s *tableState[K]) Len() int {
	return s.idx.Len()
}

func (s *tableState[K]) Total() int {
	return s.idx.Total()
}

func (s *tableState[K]) Kinds() []K {
	return s.idx.Kinds()
}

func (s *tableState[K]) Active() K {
	return s.idx.Active()
}

func (s *tableState[K]) Query() string {
	return s.idx.Query()
}

func (s *tableState[K]) Filtered() bool {
	return s.idx.Filtered()
}

func (s *tableState[K]) Searching() bool {
	return s.searching
}

func (s *tableState[K]) RefAt(i int) (Ref[K], bool) {
	return s.idx.RefAt(i)
}

func (s *tableState[K]) Cursor() int {
	return s.row
}

func (s *tableState[K]) Ref() (Ref[K], bool) {
	return s.idx.RefAt(s.row)
}

func (s *tableState[K]) SetCursor(i int) bool {
	if i < 0 || i >= s.idx.Len() {
		return false
	}

	s.row = i

	return true
}

func (s *tableState[K]) MoveUp() bool {
	if s.row <= 0 {
		return false
	}

	s.row--

	return true
}

func (s *tableState[K]) MoveDown() bool {
	if s.row >= s.idx.Len()-1 {
		return false
	}

	s.row++

	return true
}

func (s *tableState[K]) Top() {
	s.row = 0
}

func (s *tableState[K]) Bottom() {
	s.row = max(s.idx.Len()-1, 0)
}

func (s *tableState[K]) Confirming() bool {
	return s.confirming
}

func (s *tableState[K]) AskConfirm() bool {
	if s.idx.Len() == 0 {
		return false
	}

	s.confirming = true

	return true
}

func (s *tableState[K]) ClearConfirm() {
	s.confirming = false
}

func (s *tableState[K]) Help() bool {
	return s.help
}

func (s *tableState[K]) ToggleHelp() bool {
	s.help = !s.help

	return s.help
}

func (s *tableState[K]) Search(query string) {
	s.keepCursor(func() { s.idx.Search(query) })
}

func (s *tableState[K]) ClearSearch() {
	s.Search("")
}

func (s *tableState[K]) StartSearch() {
	s.searching = true
}

func (s *tableState[K]) CommitSearch() {
	s.searching = false
}

func (s *tableState[K]) CancelSearch() {
	s.searching = false
	s.ClearSearch()
}

func (s *tableState[K]) AppendSearch(text string) {
	if text == "" {
		return
	}

	s.Search(s.Query() + text)
}

func (s *tableState[K]) TrimSearch() {
	query := []rune(s.Query())
	if len(query) == 0 {
		return
	}

	s.Search(string(query[:len(query)-1]))
}

func (s *tableState[K]) SetActive(kind K) bool {
	return s.switchKind(func() bool { return s.idx.SetActive(kind) })
}

func (s *tableState[K]) NextKind() bool {
	return s.switchKind(s.idx.NextKind)
}

func (s *tableState[K]) PrevKind() bool {
	return s.switchKind(s.idx.PrevKind)
}

func (s *tableState[K]) Focus() bool {
	row, ok := s.idx.Focus(s.row)
	if !ok {
		return false
	}

	s.row = row

	return true
}

func (s *tableState[K]) Sync() {
	if s.row >= s.idx.Len() {
		s.row = max(s.idx.Len()-1, 0)
	}
	if s.row < 0 {
		s.row = 0
	}
}

func (s *tableState[K]) switchKind(step func() bool) bool {
	if !step() {
		return false
	}

	s.Top()

	return true
}

func (s *tableState[K]) keepCursor(change func()) {
	ref, ok := s.Ref()

	change()

	if ok {
		if row := slices.Index(s.idx.visible, ref); row >= 0 {
			s.row = row
			return
		}
	}

	s.Sync()
}
