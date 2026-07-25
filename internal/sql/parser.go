package sql

import (
	"errors"
	"fmt"
	"slices"
)

/*
So here I need to make a AST
After that I will also need to do the tree
executor coding
Abstract Syntax Tree

https://marianogappa.github.io/software/2019/06/05/lets-build-a-sql-parser-in-go/
*/

const (
	tokenInsertNum      = 4
	tokenDeleteNum      = 3
	tokenBasicSelectNum = 3
)

type parser struct {
	query  statement
	tokens []token
	i      int
}

type linker interface {
	exprNode()
}

type condition struct {
	inequalitySymbol TokenType
	value            string
}

type binary struct {
	operator TokenType
	r        linker
	l        linker
}

type statement struct {
	Type    TokenType
	keyType TokenType
	Key     string
	Value   string
	Where   linker
}

func (condition) exprNode() {}
func (binary) exprNode()    {}

// Connvert tokens into a statment which can be executed
func Parser(tokens []token) (statement, error) {
	p := parser{
		i:      0,
		tokens: tokens,
	}
	token := p.pop()
	if token.Type == tokEnd {
		return statement{}, errors.New("expected non empty input")
	}
	switch token.Type {
	case tokInsert:
		p.query.Type = tokInsert
		err := p.parseInsert()
		if err != nil {
			return statement{}, fmt.Errorf("failed to parse insert: %w", err)
		}
	case tokSelect:
		p.query.Type = tokSelect
		err := p.parseSelect()
		if err != nil {
			return statement{}, fmt.Errorf("failed to parse select: %w", err)
		}
	case tokDelete:
		p.query.Type = tokDelete
		err := p.parseDelete()
		if err != nil {
			return statement{}, fmt.Errorf("failed to parse delete: %w", err)
		}
	default:
		return statement{}, errors.New("synx error: invalid operation type")
	}
	if !p.isEnd() {
		return statement{}, errors.New("unexpected tokens at end")
	}
	return p.query, nil
}

// Gets the key and value from the tokens
func (p *parser) parseInsert() error {
	if len(p.tokens) != tokenInsertNum {
		return errors.New("synx error: invalid insert tokens")
	}
	keyToken := p.pop()
	valueToken := p.pop()
	if keyToken.Type != tokText || valueToken.Type != tokText {
		return errors.New("synx error: invalid insert tokens")
	}
	p.query.Key = keyToken.Value
	p.query.Value = valueToken.Value
	return nil
}

func (p *parser) parseDelete() error {
	if len(p.tokens) != tokenDeleteNum {
		return errors.New("synx error: invalid delete tokens")
	}
	keyToken := p.pop()
	if keyToken.Type != tokText {
		return errors.New("synx error: invalid delete tokens")
	}
	p.query.Key = keyToken.Value
	return nil
}

// Gets the key or all keys and the where conditions
func (p *parser) parseSelect() error {
	if len(p.tokens) < tokenBasicSelectNum {
		return errors.New("synx error: invalid select tokens")
	}
	token := p.pop()
	if token.Type != tokText && token.Type != tokAsterisk {
		return errors.New("expected valid key type")
	}
	p.query.keyType = token.Type
	p.query.Key = token.Value
	if len(p.tokens) == tokenBasicSelectNum {
		return nil
	}

	// Process where
	if !p.match(tokWhere) {
		return nil
	}

	rootExpr, err := p.orExpr()
	if err != nil {
		return fmt.Errorf("failed to get or expression: %w", err)
	}
	p.query.Where = rootExpr

	return nil
}

// Calls andExpr to get expression if there or links expression to toghter with the next
func (p *parser) orExpr() (linker, error) {
	expr, err := p.andExpr()
	if err != nil {
		return nil, fmt.Errorf("failed to get and expression: %w", err)
	}
	for p.match(tokOr) {
		right, err := p.andExpr()
		if err != nil {
			return nil, fmt.Errorf("failed to get and expression: %w", err)
		}
		expr = binary{l: expr, operator: tokOr, r: right}
	}
	return expr, nil
}

// Gets the condition and if is an and links condation toghter with and
func (p *parser) andExpr() (linker, error) {
	var expr linker
	expr, err := p.getCondition()
	if err != nil {
		return nil, fmt.Errorf("failed to get condition: %w", err)
	}
	for p.match(tokAnd) {
		right, err := p.getCondition()
		if err != nil {
			return nil, fmt.Errorf("failed to get condition: %w", err)
		}
		expr = binary{l: expr, operator: tokAnd, r: right}
	}
	return expr, err
}

func (p *parser) getCondition() (condition, error) {
	token := p.pop()
	if !slices.Contains(symbolTokens, token.Type) {
		return condition{}, errors.New("expected ineqality symbol")
	}
	inequalitySymbol := token.Type

	token = p.pop()
	if token.Type == tokEnd {
		return condition{}, fmt.Errorf("expected value")
	}
	if token.Type != tokText {
		return condition{}, errors.New("expected value to of type text")
	}
	c := condition{
		inequalitySymbol: inequalitySymbol,
		value:            token.Value,
	}
	return c, nil
}

func (p *parser) match(tok TokenType) bool {
	token := p.peek()
	if tok == token.Type && tok != tokEnd {
		p.i += 1
		return true
	}
	return false
}

// Peeks to the next token
func (p *parser) peek() token {
	return p.tokens[p.i]
}

// Get next token and move i forward until end token
func (p *parser) pop() token {
	token := p.tokens[p.i]
	if token.Type != tokEnd {
		p.i += 1
	}
	return token
}

// Check if all the tokens have been consumed or not
func (p *parser) isEnd() bool {
	return p.i == len(p.tokens)-1
}

// Used for debugging
func PrintStatement(statement statement) {
	fmt.Println("statement Type:", statement.Type)
	fmt.Println("statement keyType:", statement.keyType)
	fmt.Println("statement Key:", statement.Key)
	fmt.Println("statement value:", statement.Value)
	fmt.Println("statement where:", statement.Where)
}
