package store

import (
	"context"
)

type Storage interface {
	AddCard(ctx context.Context, columnId string, title string) (*Card, error)
	EditCard(ctx context.Context, card *Card) error
	MoveCard(ctx context.Context, toColumnId, cardId string, index int) error
	DeleteCard(ctx context.Context, columnId, cardId string, index int) error
	GetCard(ctx context.Context, cardId string) (*Card, error)
	GetCards(ctx context.Context, boardName, columnId, cardId string) ([]*Card, error)

	AddColumn(ctx context.Context, boardName, column *Column) error
	EditColumn(ctx context.Context, boardName, column *Column) error
	MoveColumn(ctx context.Context, boardName, columnId string, index uint8) error
	DeleteColumn(ctx context.Context, boardName, columnId string) error
	GetColumn(ctx context.Context, columnId string) (*Column, error)
	GetColumns(ctx context.Context, boardName string) ([]*Column, error)

	AddBoard(ctx context.Context, board *Board) error
	EditBoard(ctx context.Context, board *Board) error
	DeleteBoard(ctx context.Context, boardName string) error
	GetBoard(ctx context.Context, boardName string) (*Board, error)
}
