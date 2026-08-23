package view

import (
	"slices"
	"strings"
)

type document struct {
	id     int
	fields []string
}

type searcher interface {
	Reindex(docs []document)
	Search(query string) (hits []int, filtered bool)
}

func newSearcher() searcher {
	return &substringSearcher{}
}

type substringSearcher struct {
	docs []document
}

func (s *substringSearcher) Reindex(docs []document) {
	s.docs = make([]document, 0, len(docs))
	for _, doc := range docs {
		s.docs = append(s.docs, document{id: doc.id, fields: fold(doc.fields)})
	}
}

func (s *substringSearcher) Search(query string) ([]int, bool) {
	terms := fold(strings.Fields(query))
	if len(terms) == 0 {
		return nil, false
	}

	hits := make([]int, 0, len(s.docs))
	for _, doc := range s.docs {
		if containsAll(doc.fields, terms) {
			hits = append(hits, doc.id)
		}
	}

	return hits, true
}

func containsAll(fields, terms []string) bool {
	for _, term := range terms {
		found := slices.ContainsFunc(fields, func(field string) bool {
			return strings.Contains(field, term)
		})
		if !found {
			return false
		}
	}

	return true
}

func fold(values []string) []string {
	folded := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.ToLower(strings.TrimSpace(value)); value != "" {
			folded = append(folded, value)
		}
	}

	return folded
}

func searchable(values ...string) []string {
	fields := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			fields = append(fields, value)
		}
	}

	return fields
}
