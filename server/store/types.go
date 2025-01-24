package store

import (
	"fmt"
)

type Board struct {
	Name    string
	Columns []*Column
}

type Column struct {
	Id    string
	Index int
	Name  string
	Cards []*Card
}

type Card struct {
	Id          string
	Index       int
	Title       string
	Description string
}

type NoQueryResultsError struct{}

func (e *NoQueryResultsError) Error() string {
	return fmt.Sprintf("No results found")
}
