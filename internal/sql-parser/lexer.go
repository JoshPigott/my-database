package sqlparser

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

/*
Step there is like a lexer right
This down into key words / parts lexer spelling
Phrase
Then you must need to make a root tree (That was that stack thing)
Execution


AddToPage(key string, value string)
Delete(key string)
SelectAll() (map[string]string, error)
Select(key string) (string, bool, error)
SelectWhere(condition MathConditions, boundaryKey string) ([]data, error)

SQL commards for the thing above
AddToPage -> INSERT 'This is the key' 'This is the value'
Delete -> DELETE 'This is the key'
SelectAll -> SELECT *
Select -> SELECT 'This is the key'
Select Where -> SELECT * WHERE <= 'b'
Select Where -> SELECT * WHERE > 'b'

SQL injections

Test code
Write test but later once whole feature works
Maybe commit
*/

type TokenType string

const (
	tokAnd               TokenType = "and"
	tokOr                TokenType = "or"
	tokInsert            TokenType = "insert"
	tokSelect            TokenType = "select"
	tokDelete            TokenType = "delete"
	tokWhere             TokenType = "where"
	tokAterisk           TokenType = "*"
	tokText              TokenType = "text"
	tokNumber            TokenType = "number"
	tokComment           TokenType = "--"
	tokEqualto           TokenType = "="
	tokLessThan          TokenType = "<"
	tokLessThanOrEqualTo TokenType = "<="
	tokMoreThan          TokenType = ">"
	tokMoreThanOrEqualTo TokenType = ">="
	tokIdent             TokenType = "ident"
)

var shortCutTokens = []TokenType{tokAnd, tokOr, tokInsert, tokDelete, tokWhere, tokAterisk, tokSelect}
var symbolTokens = []TokenType{tokEqualto, tokLessThan, tokMoreThan, tokLessThanOrEqualTo, tokMoreThanOrEqualTo}
var digitChars = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

type token struct {
	tokenType TokenType
	value     string
}

// Breaks text down into tokens
func Lexer(s string) ([]token, error) {
	var err error
	var tok token
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	tokens := []token{}
	subString := ""
	i := 0
	for i < len(s) {
		char := string(s[i])

		// Comment
		if char == "-" {
			// Check if next char is "-"
			if i+1 != len(s) && string(s[i+1]) == "-" {
				i = getComment(i+2, s)
				continue
			}
		}

		// Symbol
		if isSymbol(char) && subString == "" {
			tok, i, err = getSymbol(i, s)
			if err != nil {
				return []token{}, fmt.Errorf("failed to get symbol: %w", err)
			}
			tokens = append(tokens, tok)
			continue
		}

		// String
		if char == "'" && subString == "" {
			tok, i, err = getText(i, s)
			if err != nil {
				return []token{}, fmt.Errorf("failed to get text %w", err)
			}
			tokens = append(tokens, tok)
			continue
		}

		// Digit
		if isDigit(char) && subString == "" {
			tok, i = getNumber(i, s)
			tokens = append(tokens, tok)
			continue
		}

		// White space
		if char == " " {
			found, tok := checkShortCutTokens(subString)
			if found {
				tokens = append(tokens, tok)
				subString = ""
			}

			if subString != "" {
				tokens = append(tokens, token{
					tokenType: tokIdent,
					value:     subString,
				})
				subString = ""
			}
			i += 1
			continue
		}

		subString += char
		i += 1
	}
	if subString != "" {
		found, tok := checkShortCutTokens(subString)
		if found {
			tokens = append(tokens, tok)
			subString = ""
		} else {
			tokens = append(tokens, token{
				tokenType: tokIdent,
				value:     subString,
			})
		}
	}
	return tokens, nil
}

// Get symbol and peeks ahead to check if symbol is two characters
func getSymbol(i int, s string) (token, int, error) {
	subString := string(s[i])

	isNextCharEquals := i+1 != len(s) && string(s[i+1]) == "="
	if isNextCharEquals && subString != "=" {
		subString += "="
		i += 1
	}

	for _, t := range symbolTokens {
		if string(t) == subString {
			tok := token{
				tokenType: t,
				value:     subString,
			}
			return tok, i + 1, nil
		}
	}
	return token{}, i, errors.New("no symbol found")
}

func getComment(i int, s string) int {
	// Ingore untill it hits \n
	for i < len(s) {
		char := string(s[i])
		i += 1
		if char == "\n" {
			return i
		}
	}
	return i
}

// Get the whole number and return once it is no longer a number.
func getNumber(i int, s string) (token, int) {
	var isDecimal bool
	subString := ""
	for i < len(s) {
		char := string(s[i])
		if char == "." && !isDecimal {
			isDecimal = true
			subString += char
			i += 1
		} else if isDigit(char) {
			subString += char
			i += 1
		} else {
			token := token{
				tokenType: tokNumber,
				value:     subString,
			}
			return token, i
		}
	}
	token := token{
		tokenType: tokNumber,
		value:     subString,
	}
	return token, i
}

// Get string which is wrap in single quotation marks
func getText(i int, s string) (token, int, error) {
	subString := ""
	for i < len(s) {
		char := string(s[i])
		i += 1
		// Start
		if char == "'" && subString == "" {
			continue
		}
		// End
		if char == "'" && subString != "" {
			token := token{
				tokenType: tokText,
				value:     subString,
			}
			return token, i, nil
		}
		subString += char
	}
	return token{}, 0, errors.New("syntx error string was never closed")
}

func checkShortCutTokens(subString string) (bool, token) {
	for _, t := range shortCutTokens {
		if string(t) == subString {
			return true, token{
				tokenType: t,
				value:     subString,
			}
		}
	}
	return false, token{}
}

func isDigit(char string) bool {
	return slices.Contains(digitChars, char)
}

func isSymbol(char string) bool {
	return char == "<" || char == ">" || char == "="
}
