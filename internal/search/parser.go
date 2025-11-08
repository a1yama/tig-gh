package search

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// TermType represents the type of search term
type TermType string

const (
	TermTypeKeyword TermType = "keyword"
	TermTypePhrase  TermType = "phrase"
	TermTypeAuthor  TermType = "author"
	TermTypeLabel   TermType = "label"
	TermTypeState   TermType = "state"
	TermTypeDate    TermType = "date"
)

// SearchTerm represents a single search term
type SearchTerm struct {
	Type     TermType
	Value    string
	Operator string // For date comparisons: ">", "<", ">=", "<=", "="
	Field    string // For date fields: "created", "updated"
}

// Query represents a parsed search query
type Query struct {
	Terms []SearchTerm
}

// termWithPosition tracks a term and its position in the original query
type termWithPosition struct {
	term     SearchTerm
	position int
}

// ParseQuery parses a search query string into a Query object
func ParseQuery(queryStr string) *Query {
	query := &Query{
		Terms: []SearchTerm{},
	}

	if strings.TrimSpace(queryStr) == "" {
		return query
	}

	// Regular expressions for different term types
	phraseRegex := regexp.MustCompile(`"([^"]*)"`)
	filterRegex := regexp.MustCompile(`(\w+):(>=|<=|>|<|=)?([^\s]+)`)

	termsWithPos := []termWithPosition{}

	// Extract phrases with their positions
	phraseMatches := phraseRegex.FindAllStringIndex(queryStr, -1)
	for _, match := range phraseMatches {
		phrase := strings.Trim(queryStr[match[0]:match[1]], `"`)
		if phrase != "" {
			termsWithPos = append(termsWithPos, termWithPosition{
				term: SearchTerm{
					Type:  TermTypePhrase,
					Value: phrase,
				},
				position: match[0],
			})
		}
	}

	// Extract filters with their positions
	filterMatches := filterRegex.FindAllStringSubmatchIndex(queryStr, -1)
	for _, match := range filterMatches {
		// Skip if this position is inside a phrase
		insidePhrase := false
		matchStart := match[0]
		for _, phraseMatch := range phraseMatches {
			if matchStart >= phraseMatch[0] && matchStart < phraseMatch[1] {
				insidePhrase = true
				break
			}
		}
		if insidePhrase {
			continue
		}

		field := queryStr[match[2]:match[3]]
		operator := ""
		if match[4] != -1 {
			operator = queryStr[match[4]:match[5]]
		}
		value := queryStr[match[6]:match[7]]

		var termType TermType
		switch field {
		case "author":
			termType = TermTypeAuthor
		case "label":
			termType = TermTypeLabel
		case "state":
			termType = TermTypeState
		case "created", "updated":
			termType = TermTypeDate
			if operator == "" {
				operator = "="
			}
		default:
			// Unknown filter, skip
			continue
		}

		termsWithPos = append(termsWithPos, termWithPosition{
			term: SearchTerm{
				Type:     termType,
				Value:    value,
				Operator: operator,
				Field:    field,
			},
			position: match[0],
		})
	}

	// Mark positions that have been matched
	matched := make([]bool, len(queryStr))
	// Mark phrase positions
	for _, match := range phraseMatches {
		for i := match[0]; i < match[1]; i++ {
			matched[i] = true
		}
	}
	// Mark filter positions
	for _, match := range filterMatches {
		for i := match[0]; i < match[1]; i++ {
			matched[i] = true
		}
	}

	// Extract keywords (unmatched words)
	currentWord := ""
	wordStart := 0
	for i := 0; i < len(queryStr); i++ {
		if !matched[i] && queryStr[i] != ' ' && queryStr[i] != '\t' {
			if currentWord == "" {
				wordStart = i
			}
			currentWord += string(queryStr[i])
		} else {
			if currentWord != "" {
				termsWithPos = append(termsWithPos, termWithPosition{
					term: SearchTerm{
						Type:  TermTypeKeyword,
						Value: currentWord,
					},
					position: wordStart,
				})
				currentWord = ""
			}
		}
	}
	// Don't forget the last word
	if currentWord != "" {
		termsWithPos = append(termsWithPos, termWithPosition{
			term: SearchTerm{
				Type:  TermTypeKeyword,
				Value: currentWord,
			},
			position: wordStart,
		})
	}

	// Sort by position
	sort.Slice(termsWithPos, func(i, j int) bool {
		return termsWithPos[i].position < termsWithPos[j].position
	})

	// Extract terms in order
	for _, twp := range termsWithPos {
		query.Terms = append(query.Terms, twp.term)
	}

	return query
}

// HasFilter checks if the query has a filter of the given type
func (q *Query) HasFilter(termType TermType) bool {
	for _, term := range q.Terms {
		if term.Type == termType {
			return true
		}
	}
	return false
}

// GetFilter returns all terms of the given type
func (q *Query) GetFilter(termType TermType) []SearchTerm {
	var results []SearchTerm
	for _, term := range q.Terms {
		if term.Type == termType {
			results = append(results, term)
		}
	}
	return results
}

// GetKeywords returns all keyword terms
func (q *Query) GetKeywords() []string {
	var keywords []string
	for _, term := range q.Terms {
		if term.Type == TermTypeKeyword {
			keywords = append(keywords, term.Value)
		}
	}
	return keywords
}

// GetPhrases returns all phrase terms
func (q *Query) GetPhrases() []string {
	var phrases []string
	for _, term := range q.Terms {
		if term.Type == TermTypePhrase {
			phrases = append(phrases, term.Value)
		}
	}
	return phrases
}

// IsEmpty returns true if the query has no terms
func (q *Query) IsEmpty() bool {
	return len(q.Terms) == 0
}

// String returns a string representation of the query
func (q *Query) String() string {
	if q.IsEmpty() {
		return ""
	}

	var parts []string
	for _, term := range q.Terms {
		switch term.Type {
		case TermTypePhrase:
			parts = append(parts, fmt.Sprintf(`"%s"`, term.Value))
		case TermTypeAuthor:
			parts = append(parts, fmt.Sprintf("author:%s", term.Value))
		case TermTypeLabel:
			parts = append(parts, fmt.Sprintf("label:%s", term.Value))
		case TermTypeState:
			parts = append(parts, fmt.Sprintf("state:%s", term.Value))
		case TermTypeDate:
			if term.Operator != "" && term.Operator != "=" {
				parts = append(parts, fmt.Sprintf("%s:%s%s", term.Field, term.Operator, term.Value))
			} else {
				parts = append(parts, fmt.Sprintf("%s:%s", term.Field, term.Value))
			}
		case TermTypeKeyword:
			parts = append(parts, term.Value)
		}
	}
	return strings.Join(parts, " ")
}

// ParseDate parses the date value of a date term
func (t *SearchTerm) ParseDate() (time.Time, error) {
	if t.Type != TermTypeDate {
		return time.Time{}, fmt.Errorf("term is not a date type")
	}

	// Try different date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
	}

	var lastErr error
	for _, format := range formats {
		parsed, err := time.Parse(format, t.Value)
		if err == nil {
			return parsed, nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("invalid date format: %w", lastErr)
}

// MatchDate checks if the given date matches the term's criteria
func (t *SearchTerm) MatchDate(date time.Time) bool {
	if t.Type != TermTypeDate {
		return false
	}

	termDate, err := t.ParseDate()
	if err != nil {
		return false
	}

	switch t.Operator {
	case ">":
		return date.After(termDate)
	case ">=":
		return date.After(termDate) || date.Equal(termDate)
	case "<":
		return date.Before(termDate)
	case "<=":
		return date.Before(termDate) || date.Equal(termDate)
	case "=", "":
		// Compare just the date part (ignore time)
		y1, m1, d1 := date.Date()
		y2, m2, d2 := termDate.Date()
		return y1 == y2 && m1 == m2 && d1 == d2
	default:
		return false
	}
}
