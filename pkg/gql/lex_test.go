package gql

import (
	"fmt"
	"strings"
	"testing"
	"text/scanner"
	"unicode"
)

var _gqlTestSample = `
query HeroComparison($first: Int = 3) {
	leftComparison: hero(episode: EMPIRE) {
	  ...comparisonFields
	}
	rightComparison: hero(episode: JEDI) {
	  ...comparisonFields
	}
  }
`

func TestLex(t *testing.T) {

	var s scanner.Scanner

	s.IsIdentRune = func(ch rune, i int) bool {
		return ch == '$' && i == 0 || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
	}

	r := strings.NewReader(_gqlTestSample)

	s.Init(r)
	s.Filename = "default"

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		fmt.Printf("%s: %s\n", s.Position, s.TokenText())
	}
}
