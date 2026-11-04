package ui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/view"
)

const modalMaxInner = 120

const footerLines = 2

type screenSize struct {
	width, height int
}

type focusSpan struct {
	from, row, to int
}

func focusRow(row int) focusSpan {
	return focusSpan{from: row, row: row, to: row}
}

func (s *screenSize) observe(msg tea.Msg) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		s.width, s.height = size.Width, size.Height
	}
}

func (s screenSize) frameWidth(minInner int) int {
	return max(s.width-2, minInner)
}

func (s screenSize) extra(minInner int) int {
	return s.frameWidth(minInner) - minInner
}

func (s screenSize) modalWidth(minInner int) int {
	return min(s.frameWidth(minInner), max(modalMaxInner, minInner))
}

func (s screenSize) formInner() int {
	return s.modalWidth(formInner)
}

func (s screenSize) formBox() int {
	return formBoxW + s.formInner() - formInner
}

func (s screenSize) follow(rows *view.Scroll, chrome int, focus focusSpan, total int) {
	if s.height > 0 {
		rows.Resize(max(s.height-chrome, 1))
	}
	rows.Follow(focus.to, total)
	rows.Follow(focus.from, total)
	rows.Follow(focus.row, total)
}

func (s screenSize) visibleRows(rows *view.Scroll, chrome, cursor, total int) (int, int) {
	start, end := rows.Range(total)
	if s.height == 0 {
		return start, end
	}
	keep := min(end-start, max(s.height-chrome, 1))
	from := min(max(start, cursor-keep+1), end-keep)
	return from, from + keep
}

func (s screenSize) fit(rows *view.Scroll, chrome, cursor, total int, height func() int) {
	if s.height == 0 {
		return
	}
	budget := s.height - chrome
	settled := *rows
	for size := min(total, max(budget, 1)); size >= 1; size-- {
		*rows = settled
		rows.Resize(size)
		rows.Follow(cursor, total)
		if height() <= budget {
			return
		}
	}
}

func (s screenSize) keepEnd(lines []string, chrome int) []string {
	if s.height == 0 {
		return lines
	}
	return lines[max(len(lines)-max(s.height-chrome, 1), 0):]
}

func (s screenSize) join(lines []string) string {
	if s.height > footerLines && len(lines) > s.height {
		kept := slices.Clone(lines[:s.height-footerLines])
		lines = append(kept, lines[len(lines)-footerLines:]...)
	}
	return strings.Join(lines, "\n")
}

type follower[M any] interface {
	tea.Model
	followCursor() M
}

func followed[M follower[M]](next tea.Model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if m, ok := next.(M); ok {
		return m.followCursor(), cmd
	}
	return next, cmd
}

func topNote(parts ...string) string {
	var kept []string
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	if len(kept) == 0 {
		return ""
	}
	return " " + strings.Join(kept, " · ") + " " + gLineH
}
