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

type NotFoundError struct {
	typ string
	id  string
}

func NewNotFoundError(typ, id string) *NotFoundError {
	return &NotFoundError{
		typ: typ,
		id:  id,
	}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %s not found", e.typ, e.id)
}

func (e NotFoundError) Is(target error) bool {
	// Check if the target is a NotFoundError, and if so, compare some fields
	if targetErr, ok := target.(*NotFoundError); ok {
		return e.typ == targetErr.typ && e.id == targetErr.id
	}
	return false
}

type BadRequestError struct {
	issue string
}

func NewBadRequestError(issue string) *BadRequestError {
	return &BadRequestError{
		issue: issue,
	}
}

func (e BadRequestError) Error() string {
	return e.issue
}
