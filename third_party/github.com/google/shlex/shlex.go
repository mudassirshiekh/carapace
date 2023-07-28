/*
Copyright 2012 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package shlex implements a simple lexer which splits input in to tokens using
shell-style rules for quoting and commenting.

The basic use case uses the default ASCII lexer to split a string into sub-strings:

	shlex.Split("one \"two three\" four") -> []string{"one", "two three", "four"}

To process a stream of strings:

	l := NewLexer(os.Stdin)
	for ; token, err := l.Next(); err != nil {
		// process token
	}

To access the raw token stream (which includes tokens for comments):

	  t := NewTokenizer(os.Stdin)
	  for ; token, err := t.Next(); err != nil {
		// process token
	  }
*/
package shlex

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// TokenType is a top-level token classification: A word, space, comment, unknown.
type TokenType int

// runeTokenClass is the type of a UTF-8 character classification: A quote, space, escape.
type runeTokenClass int

// the internal state used by the lexer state machine
type lexerState int

// Token is a (type, value) pair representing a lexographical token.
type Token struct {
	tokenType TokenType
	value     string
	index     int
}

// Equal reports whether tokens a, and b, are equal.
// Two tokens are equal if both their types and values are equal. A nil token can
// never be equal to another token.
func (a *Token) Equal(b *Token) bool {
	if a == nil || b == nil {
		return false
	}
	if a.tokenType != b.tokenType {
		return false
	}
	return a.value == b.value && a.index == b.index
}

// Named classes of UTF-8 runes
const (
	spaceRunes            = " \t\r\n"
	escapingQuoteRunes    = `"`
	nonEscapingQuoteRunes = "'"
	escapeRunes           = `\`
	commentRunes          = "#"
	pipelineRunes         = "|&;"
)

// Classes of rune token
const (
	unknownRuneClass runeTokenClass = iota
	spaceRuneClass
	escapingQuoteRuneClass
	nonEscapingQuoteRuneClass
	escapeRuneClass
	commentRuneClass
	pipelineRuneClass
	eofRuneClass
)

// Classes of lexographic token
const (
	UnknownToken TokenType = iota
	WordToken
	SpaceToken
	CommentToken
	PipelineToken
)

// Lexer state machine states
const (
	startState           lexerState = iota // no runes have been seen
	inWordState                            // processing regular runes in a word
	escapingState                          // we have just consumed an escape rune; the next rune is literal
	escapingQuotedState                    // we have just consumed an escape rune within a quoted string
	quotingEscapingState                   // we are within a quoted string that supports escaping ("...")
	quotingState                           // we are within a string that does not support escaping ('...')
	commentState                           // we are within a comment (everything following an unquoted or unescaped #
)

// tokenClassifier is used for classifying rune characters.
type tokenClassifier map[rune]runeTokenClass

func (typeMap tokenClassifier) addRuneClass(runes string, tokenType runeTokenClass) {
	for _, runeChar := range runes {
		typeMap[runeChar] = tokenType
	}
}

// newDefaultClassifier creates a new classifier for ASCII characters.
func newDefaultClassifier() tokenClassifier {
	t := tokenClassifier{}
	t.addRuneClass(spaceRunes, spaceRuneClass)
	t.addRuneClass(escapingQuoteRunes, escapingQuoteRuneClass)
	t.addRuneClass(nonEscapingQuoteRunes, nonEscapingQuoteRuneClass)
	t.addRuneClass(escapeRunes, escapeRuneClass)
	t.addRuneClass(commentRunes, commentRuneClass)
	t.addRuneClass(pipelineRunes, pipelineRuneClass)
	return t
}

// ClassifyRune classifiees a rune
func (t tokenClassifier) ClassifyRune(runeVal rune) runeTokenClass {
	return t[runeVal]
}

// Lexer turns an input stream into a sequence of tokens. Whitespace and comments are skipped.
type Lexer Tokenizer

// NewLexer creates a new lexer from an input stream.
func NewLexer(r io.Reader) *Lexer {

	return (*Lexer)(NewTokenizer(r))
}

type PipelineSeparatorError struct{}

func (m *PipelineSeparatorError) Error() string {
	return "encountered a pipeline separator like `|`"
}

// Next returns the next word, or an error. If there are no more words,
// the error will be io.EOF.
func (l *Lexer) Next() (string, int, error) {
	for {
		token, err := (*Tokenizer)(l).Next()
		if err != nil {
			return "", -1, err
		}
		switch token.tokenType {
		case WordToken:
			return token.value, token.index, nil
		case CommentToken:
			// skip comments
		case PipelineToken:
			// return token but with pseudo err to mark end of pipeline
			return token.value, token.index, &PipelineSeparatorError{}
		default:
			return "", token.index, fmt.Errorf("Unknown token type: %v", token.tokenType)
		}
	}
}

// Tokenizer turns an input stream into a sequence of typed tokens
type Tokenizer struct {
	input      bufio.Reader
	classifier tokenClassifier
	index      int
}

// NewTokenizer creates a new tokenizer from an input stream.
func NewTokenizer(r io.Reader) *Tokenizer {
	input := bufio.NewReader(r)
	classifier := newDefaultClassifier()
	return &Tokenizer{
		input:      *input,
		classifier: classifier}
}

// scanStream scans the stream for the next token using the internal state machine.
// It will panic if it encounters a rune which it does not know how to handle.
func (t *Tokenizer) scanStream() (*Token, error) {
	state := startState
	tokenIndex := 0
	var tokenType TokenType
	var value []rune
	var nextRune rune
	var nextRuneType runeTokenClass
	var err error

	for {
		nextRune, _, err = t.input.ReadRune()
		t.index += 1
		nextRuneType = t.classifier.ClassifyRune(nextRune)

		if err == io.EOF {
			nextRuneType = eofRuneClass
			err = nil
		} else if err != nil {
			return nil, err
		}

		switch state {
		case startState: // no runes read yet
			{
				if nextRuneType != spaceRuneClass {
					tokenIndex = t.index - 1 // TODO verify
				}
				switch nextRuneType {
				case eofRuneClass:
					{
						return nil, io.EOF
					}
				case spaceRuneClass:
					{
					}
				case escapingQuoteRuneClass:
					{
						tokenType = WordToken
						state = quotingEscapingState
					}
				case nonEscapingQuoteRuneClass:
					{
						tokenType = WordToken
						state = quotingState
					}
				case escapeRuneClass:
					{
						tokenType = WordToken
						state = escapingState
					}
				case commentRuneClass:
					{
						tokenType = CommentToken
						state = commentState
					}
				case pipelineRuneClass:
					{
						tokenType = PipelineToken
						value = append(value, nextRune)
						state = inWordState
					}
				default:
					{
						tokenType = WordToken
						value = append(value, nextRune)
						state = inWordState
					}
				}
			}
		case inWordState: // in a regular word
			{
				switch nextRuneType {
				case eofRuneClass:
					{
						token := &Token{
							tokenType: tokenType,
							value:     string(value),
							index:     tokenIndex}
						return token, err
					}
				case spaceRuneClass:
					{
						token := &Token{
							tokenType: tokenType,
							value:     string(value),
							index:     tokenIndex}
						return token, err
					}
				case escapingQuoteRuneClass:
					{
						state = quotingEscapingState
					}
				case nonEscapingQuoteRuneClass:
					{
						state = quotingState
					}
				case escapeRuneClass:
					{
						state = escapingState
					}
				default:
					{
						value = append(value, nextRune)
					}
				}
			}
		case escapingState: // the rune after an escape character
			{
				switch nextRuneType {
				case eofRuneClass:
					{
						err = fmt.Errorf("EOF found after escape character")
						token := &Token{
							tokenType: tokenType,
							value:     string(value),
							index:     tokenIndex}
						return token, err
					}
				default:
					{
						state = inWordState
						value = append(value, nextRune)
					}
				}
			}
		case escapingQuotedState: // the next rune after an escape character, in double quotes
			{
				switch nextRuneType {
				case eofRuneClass:
					{
						err = fmt.Errorf("EOF found after escape character")
						token := &Token{
							tokenType: tokenType,
							value:     string(value),
							index:     tokenIndex}
						return token, err
					}
				default:
					{
						state = quotingEscapingState
						value = append(value, nextRune)
					}
				}
			}
		case quotingEscapingState: // in escaping double quotes
			{
				switch nextRuneType {
				case eofRuneClass:
					{
						err = fmt.Errorf("EOF found when expecting closing quote")
						token := &Token{
							tokenType: tokenType,
							value:     string(value),
							index:     tokenIndex}
						return token, err
					}
				case escapingQuoteRuneClass:
					{
						state = inWordState
					}
				case escapeRuneClass:
					{
						state = escapingQuotedState
					}
				default:
					{
						value = append(value, nextRune)
					}
				}
			}
		case quotingState: // in non-escaping single quotes
			{
				switch nextRuneType {
				case eofRuneClass:
					{
						err = fmt.Errorf("EOF found when expecting closing quote")
						token := &Token{
							tokenType: tokenType,
							value:     string(value),
							index:     tokenIndex}
						return token, err
					}
				case nonEscapingQuoteRuneClass:
					{
						state = inWordState
					}
				default:
					{
						value = append(value, nextRune)
					}
				}
			}
		case commentState: // in a comment
			{
				switch nextRuneType {
				case eofRuneClass:
					{
						token := &Token{
							tokenType: tokenType,
							value:     string(value),
							index:     tokenIndex}
						return token, err
					}
				case spaceRuneClass:
					{
						if nextRune == '\n' {
							state = startState
							token := &Token{
								tokenType: tokenType,
								value:     string(value),
								index:     tokenIndex}
							return token, err
						} else {
							value = append(value, nextRune)
						}
					}
				default:
					{
						value = append(value, nextRune)
					}
				}
			}
		default:
			{
				return nil, fmt.Errorf("Unexpected state: %v", state)
			}
		}
	}
}

// Next returns the next token in the stream.
func (t *Tokenizer) Next() (*Token, error) {
	return t.scanStream()
}

// Split partitions a string into a slice of strings.
func Split(s string) ([]string, error) {
	substrings, _, err := SplitP(s, false)
	return substrings, err
}

// SplitP is like Split but supports pipelines and returns the prefix.
func SplitP(s string, pipelines bool) ([]string, string, error) {
	l := NewLexer(strings.NewReader(s))
	subStrings := make([]string, 0)
	lastIndex := 0
	for {
		word, index, err := l.Next()
		if err != nil {
			if err == io.EOF {
				return subStrings, string([]rune(s)[:lastIndex]), nil
			}
			if _, ok := err.(*PipelineSeparatorError); !ok {
				return subStrings, "", err
			}
			if pipelines {
				subStrings = make([]string, 0)
				lastIndex = index
				continue
			}
		}
		subStrings = append(subStrings, word)
		lastIndex = index
	}
}
