package sql

import (
	"bubbly-database/internal/database"
	"errors"
	"fmt"
)

// Complies query into statement then execute statement.
func Query(DB *database.DB, s string) error {
	statement, err := compileQuery(s)
	if err != nil {
		return fmt.Errorf("failed to compile query: %w", err)
	}
	if err := statement.execute(DB); err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}
	return nil
}

// Same as query just returns data.
func QuerySelect(DB *database.DB, s string) ([]database.Data, error) {
	statement, err := compileQuery(s)
	if err != nil {
		return []database.Data{}, fmt.Errorf("failed to compile query: %w", err)
	}
	data, err := statement.executeSelect(DB)
	if err != nil {
		return []database.Data{}, fmt.Errorf("failed to execute statement: %w", err)
	}
	return data, nil
}

// Turns input string into tokens with lexer, then tokens into statement with parser.
func compileQuery(s string) (statement, error) {
	tokens, err := lexer(s)
	if err != nil {
		return statement{}, fmt.Errorf("failed to get tokens with lexer %w", err)
	}
	stmt, err := parse(tokens)
	if err != nil {
		return statement{}, fmt.Errorf("failed to parse tokens into statement: %w", err)
	}
	return stmt, nil
}

func (s statement) execute(DB *database.DB) error {
	switch s.Type {
	case tokInsert:
		err := s.executeInsert(DB)
		if err != nil {
			return fmt.Errorf("failed to execute insert: %w", err)
		}

	case tokDelete:
		err := s.executeDelete(DB)
		if err != nil {
			return fmt.Errorf("failed to execute delete: %w", err)
		}
	default:
		return errors.New("unrecognised statement")
	}
	return nil
}

// Decides which funcation to call selectALL or Select or SelectWhere.
func (s statement) executeSelect(DB *database.DB) ([]database.Data, error) {
	if s.Where == nil {
		if s.keyType == tokAsterisk {
			data, err := DB.SelectAll()
			if err != nil {
				return []database.Data{}, fmt.Errorf("failed to select all data: %w", err)
			}
			return data, err
		}
		if s.keyType == tokText {
			data, _, err := DB.Select(s.Key)
			if err != nil {
				return []database.Data{}, fmt.Errorf("failed to select data with key %q: %w", s.Key, err)
			}
			return data, err
		}
		return []database.Data{}, errors.New("unrecognised statement")
	}
	c, ok := s.Where.(condition)
	if !ok {
		return []database.Data{}, errors.New("logic expression and multiple conditions not supported")
	}
	data, err := DB.SelectWhere(database.MathConditions(c.inequalitySymbol), c.value)
	if err != nil {
		return []database.Data{}, fmt.Errorf("failed to select where %s %q: %w", c.inequalitySymbol, c.value, err)
	}
	return data, nil
}

func (s statement) executeInsert(DB *database.DB) error {
	if err := DB.AddToPage(s.Key, s.Value); err != nil {
		return fmt.Errorf("failed to add to page: %w", err)
	}
	return nil
}

func (s statement) executeDelete(DB *database.DB) error {
	if err := DB.Delete(s.Key); err != nil {
		return fmt.Errorf("failed to delete %v: %w", s.Key, err)
	}
	return nil
}
