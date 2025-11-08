package search

import (
	"reflect"
	"testing"
	"time"
)

func TestParseQuery_EmptyString(t *testing.T) {
	query := ParseQuery("")

	if query == nil {
		t.Fatal("ParseQuery should not return nil")
	}

	if len(query.Terms) != 0 {
		t.Errorf("Expected 0 terms for empty query, got %d", len(query.Terms))
	}
}

func TestParseQuery_SimpleKeyword(t *testing.T) {
	query := ParseQuery("bug")

	if len(query.Terms) != 1 {
		t.Fatalf("Expected 1 term, got %d", len(query.Terms))
	}

	term := query.Terms[0]
	if term.Type != TermTypeKeyword {
		t.Errorf("Expected type %s, got %s", TermTypeKeyword, term.Type)
	}

	if term.Value != "bug" {
		t.Errorf("Expected value 'bug', got %q", term.Value)
	}
}

func TestParseQuery_MultipleKeywords(t *testing.T) {
	query := ParseQuery("bug fix login")

	if len(query.Terms) != 3 {
		t.Fatalf("Expected 3 terms, got %d", len(query.Terms))
	}

	expectedValues := []string{"bug", "fix", "login"}
	for i, expected := range expectedValues {
		if query.Terms[i].Value != expected {
			t.Errorf("Expected term %d to be %q, got %q", i, expected, query.Terms[i].Value)
		}
	}
}

func TestParseQuery_AuthorFilter(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple author", "author:john", "john"},
		{"author with underscore", "author:john_doe", "john_doe"},
		{"author with hyphen", "author:john-doe", "john-doe"},
		{"author with numbers", "author:user123", "user123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := ParseQuery(tc.input)

			if len(query.Terms) != 1 {
				t.Fatalf("Expected 1 term, got %d", len(query.Terms))
			}

			term := query.Terms[0]
			if term.Type != TermTypeAuthor {
				t.Errorf("Expected type %s, got %s", TermTypeAuthor, term.Type)
			}

			if term.Value != tc.expected {
				t.Errorf("Expected value %q, got %q", tc.expected, term.Value)
			}
		})
	}
}

func TestParseQuery_LabelFilter(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple label", "label:bug", "bug"},
		{"label with hyphen", "label:priority-high", "priority-high"},
		{"label with underscore", "label:good_first_issue", "good_first_issue"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := ParseQuery(tc.input)

			if len(query.Terms) != 1 {
				t.Fatalf("Expected 1 term, got %d", len(query.Terms))
			}

			term := query.Terms[0]
			if term.Type != TermTypeLabel {
				t.Errorf("Expected type %s, got %s", TermTypeLabel, term.Type)
			}

			if term.Value != tc.expected {
				t.Errorf("Expected value %q, got %q", tc.expected, term.Value)
			}
		})
	}
}

func TestParseQuery_StateFilter(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"open state", "state:open", "open"},
		{"closed state", "state:closed", "closed"},
		{"all state", "state:all", "all"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := ParseQuery(tc.input)

			if len(query.Terms) != 1 {
				t.Fatalf("Expected 1 term, got %d", len(query.Terms))
			}

			term := query.Terms[0]
			if term.Type != TermTypeState {
				t.Errorf("Expected type %s, got %s", TermTypeState, term.Type)
			}

			if term.Value != tc.expected {
				t.Errorf("Expected value %q, got %q", tc.expected, term.Value)
			}
		})
	}
}

func TestParseQuery_DateFilter(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		operator string
		date     string
	}{
		{"greater than", "created:>2024-01-01", ">", "2024-01-01"},
		{"less than", "created:<2024-12-31", "<", "2024-12-31"},
		{"equal", "created:2024-06-15", "=", "2024-06-15"},
		{"greater or equal", "updated:>=2024-01-01", ">=", "2024-01-01"},
		{"less or equal", "updated:<=2024-12-31", "<=", "2024-12-31"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := ParseQuery(tc.input)

			if len(query.Terms) != 1 {
				t.Fatalf("Expected 1 term, got %d", len(query.Terms))
			}

			term := query.Terms[0]
			if term.Type != TermTypeDate {
				t.Errorf("Expected type %s, got %s", TermTypeDate, term.Type)
			}

			if term.Operator != tc.operator {
				t.Errorf("Expected operator %q, got %q", tc.operator, term.Operator)
			}

			// Value should contain the date
			if term.Value != tc.date {
				t.Errorf("Expected value %q, got %q", tc.date, term.Value)
			}
		})
	}
}

func TestParseQuery_PhraseSearch(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple phrase", `"bug fix"`, "bug fix"},
		{"phrase with special chars", `"login failed error"`, "login failed error"},
		{"phrase with punctuation", `"can't login"`, "can't login"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := ParseQuery(tc.input)

			if len(query.Terms) != 1 {
				t.Fatalf("Expected 1 term, got %d", len(query.Terms))
			}

			term := query.Terms[0]
			if term.Type != TermTypePhrase {
				t.Errorf("Expected type %s, got %s", TermTypePhrase, term.Type)
			}

			if term.Value != tc.expected {
				t.Errorf("Expected value %q, got %q", tc.expected, term.Value)
			}
		})
	}
}

func TestParseQuery_MixedQuery(t *testing.T) {
	query := ParseQuery(`author:john label:bug state:open "login error" fix`)

	expectedTerms := []struct {
		termType TermType
		value    string
	}{
		{TermTypeAuthor, "john"},
		{TermTypeLabel, "bug"},
		{TermTypeState, "open"},
		{TermTypePhrase, "login error"},
		{TermTypeKeyword, "fix"},
	}

	if len(query.Terms) != len(expectedTerms) {
		t.Fatalf("Expected %d terms, got %d", len(expectedTerms), len(query.Terms))
	}

	for i, expected := range expectedTerms {
		term := query.Terms[i]
		if term.Type != expected.termType {
			t.Errorf("Term %d: expected type %s, got %s", i, expected.termType, term.Type)
		}
		if term.Value != expected.value {
			t.Errorf("Term %d: expected value %q, got %q", i, expected.value, term.Value)
		}
	}
}

func TestParseQuery_ExtraWhitespace(t *testing.T) {
	query := ParseQuery("  bug    fix   ")

	if len(query.Terms) != 2 {
		t.Fatalf("Expected 2 terms, got %d", len(query.Terms))
	}

	if query.Terms[0].Value != "bug" {
		t.Errorf("Expected first term to be 'bug', got %q", query.Terms[0].Value)
	}

	if query.Terms[1].Value != "fix" {
		t.Errorf("Expected second term to be 'fix', got %q", query.Terms[1].Value)
	}
}

func TestParseQuery_InvalidDateFormat(t *testing.T) {
	// Should handle gracefully, possibly as keyword or ignore
	query := ParseQuery("created:invalid-date")

	// Implementation may vary - just ensure no panic
	if query == nil {
		t.Error("ParseQuery should not return nil even for invalid input")
	}
}

func TestQuery_HasFilter(t *testing.T) {
	query := ParseQuery("author:john label:bug keyword")

	if !query.HasFilter(TermTypeAuthor) {
		t.Error("Expected query to have author filter")
	}

	if !query.HasFilter(TermTypeLabel) {
		t.Error("Expected query to have label filter")
	}

	if query.HasFilter(TermTypeState) {
		t.Error("Expected query not to have state filter")
	}
}

func TestQuery_GetFilter(t *testing.T) {
	query := ParseQuery("author:john author:jane label:bug")

	// Get author filters
	authorTerms := query.GetFilter(TermTypeAuthor)
	if len(authorTerms) != 2 {
		t.Fatalf("Expected 2 author filters, got %d", len(authorTerms))
	}

	// Get label filters
	labelTerms := query.GetFilter(TermTypeLabel)
	if len(labelTerms) != 1 {
		t.Fatalf("Expected 1 label filter, got %d", len(labelTerms))
	}

	// Get non-existent filter
	stateTerms := query.GetFilter(TermTypeState)
	if len(stateTerms) != 0 {
		t.Errorf("Expected 0 state filters, got %d", len(stateTerms))
	}
}

func TestQuery_GetKeywords(t *testing.T) {
	query := ParseQuery("bug fix author:john critical")

	keywords := query.GetKeywords()
	expected := []string{"bug", "fix", "critical"}

	if !reflect.DeepEqual(keywords, expected) {
		t.Errorf("Expected keywords %v, got %v", expected, keywords)
	}
}

func TestQuery_GetPhrases(t *testing.T) {
	query := ParseQuery(`"login error" bug "cannot authenticate"`)

	phrases := query.GetPhrases()
	expected := []string{"login error", "cannot authenticate"}

	if !reflect.DeepEqual(phrases, expected) {
		t.Errorf("Expected phrases %v, got %v", expected, phrases)
	}
}

func TestSearchTerm_MatchDate(t *testing.T) {
	testDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		operator string
		value    string
		matches  bool
	}{
		{"greater than - match", ">", "2024-01-01", true},
		{"greater than - no match", ">", "2024-12-31", false},
		{"less than - match", "<", "2024-12-31", true},
		{"less than - no match", "<", "2024-01-01", false},
		{"equal - match", "=", "2024-06-15", true},
		{"equal - no match", "=", "2024-06-14", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			term := &SearchTerm{
				Type:     TermTypeDate,
				Operator: tc.operator,
				Value:    tc.value,
			}

			result := term.MatchDate(testDate)
			if result != tc.matches {
				t.Errorf("Expected match=%v, got %v", tc.matches, result)
			}
		})
	}
}

func TestSearchTerm_ParseDate(t *testing.T) {
	term := &SearchTerm{
		Type:  TermTypeDate,
		Value: "2024-06-15",
	}

	parsedDate, err := term.ParseDate()
	if err != nil {
		t.Fatalf("Unexpected error parsing date: %v", err)
	}

	expected := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if !parsedDate.Equal(expected) {
		t.Errorf("Expected date %v, got %v", expected, parsedDate)
	}
}

func TestSearchTerm_ParseDate_Invalid(t *testing.T) {
	term := &SearchTerm{
		Type:  TermTypeDate,
		Value: "invalid",
	}

	_, err := term.ParseDate()
	if err == nil {
		t.Error("Expected error for invalid date format")
	}
}

func TestQuery_String(t *testing.T) {
	query := ParseQuery(`author:john label:bug "login error"`)

	str := query.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	// Should contain the original terms in some form
	if str == "" {
		t.Error("String() should return a representation of the query")
	}
}

func TestParseQuery_CaseSensitivity(t *testing.T) {
	// Keywords should preserve case
	query := ParseQuery("Bug FIX")

	if len(query.Terms) != 2 {
		t.Fatalf("Expected 2 terms, got %d", len(query.Terms))
	}

	if query.Terms[0].Value != "Bug" {
		t.Errorf("Expected 'Bug', got %q", query.Terms[0].Value)
	}

	if query.Terms[1].Value != "FIX" {
		t.Errorf("Expected 'FIX', got %q", query.Terms[1].Value)
	}
}

func TestParseQuery_UnterminatedQuote(t *testing.T) {
	// Should handle gracefully
	query := ParseQuery(`"unterminated quote`)

	// Implementation may vary - ensure no panic
	if query == nil {
		t.Error("ParseQuery should not return nil")
	}
}

func TestQuery_IsEmpty(t *testing.T) {
	emptyQuery := ParseQuery("")
	if !emptyQuery.IsEmpty() {
		t.Error("Expected empty query to return true for IsEmpty()")
	}

	nonEmptyQuery := ParseQuery("bug")
	if nonEmptyQuery.IsEmpty() {
		t.Error("Expected non-empty query to return false for IsEmpty()")
	}
}
