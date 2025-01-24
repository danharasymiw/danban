package mdb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/danharasymiw/danban/server/store"
)

type MongoDb struct {
	client    *mongo.Client
	boardCol  *mongo.Collection
	columnCol *mongo.Collection
	cardCol   *mongo.Collection
}

const dbName = "danban"

func New() *MongoDb {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	boardCol := client.Database(dbName).Collection("boards")
	columnCol := client.Database(dbName).Collection("columns")
	cardCol := client.Database(dbName).Collection("cards")

	return &MongoDb{
		client:    client,
		boardCol:  boardCol,
		columnCol: columnCol,
		cardCol:   cardCol,
	}
}

func (m *MongoDb) AddCard(ctx context.Context, boardId, columnId string, card *store.Card) error {

	return errors.New(`Not implemented`)
}

func (m *MongoDb) EditCard(ctx context.Context, boardId, columnId, card *store.Card) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) MoveCard(ctx context.Context, boardId, columnId string, index uint8) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) DeleteCard(ctx context.Context, boardId, columnId, cardId string) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) GetCard(ctx context.Context, boardId, columnId, cardId string) (*store.Card, error) {
	return nil, errors.New(`Not implemented`)
}

func (m *MongoDb) GetCards(ctx context.Context, boardId, columnId, cardId string) ([]*store.Card, error) {
	return nil, errors.New(`Not implemented`)
}

func (m *MongoDb) AddColumn(ctx context.Context, boardId, column *store.Column) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) EditColumn(ctx context.Context, boardId, column *store.Column) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) MoveColumn(ctx context.Context, boardId, columnId string, index uint8) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) DeleteColumn(ctx context.Context, boardId, columnId string) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) GetColumn(ctx context.Context, boardId, columnId string) (*store.Column, error) {
	return nil, errors.New(`Not implemented`)
}

func (m *MongoDb) GetColumns(ctx context.Context, boardId, columnId string) ([]*store.Column, error) {
	return nil, errors.New(`Not implemented`)
}

func (m *MongoDb) AddBoard(ctx context.Context, boardDTO *store.Board) error {
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("could not start session: %v", err)
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		var columnIds []primitive.ObjectID
		for _, column := range boardDTO.Columns {
			colRes, err := m.columnCol.InsertOne(sc, column)
			if err != nil {
				return fmt.Errorf("could not insert column: %v", err)
			}
			colId := colRes.InsertedID.(primitive.ObjectID)
			column.Id = colId.Hex()
			columnIds = append(columnIds, colId)

			for _, card := range column.Cards {
				newCard := &Card{
					Title:       card.Title,
					Description: card.Description,
					Index:       card.Index,
					ColumnId:    colId,
				}
				cardRes, err := m.cardCol.InsertOne(sc, newCard)
				if err != nil {
					return fmt.Errorf("could not insert card: %v", err)
				}
				card.Id = cardRes.InsertedID.(primitive.ObjectID).Hex()
			}
		}

		newBoard := &Board{
			Name:      boardDTO.Name,
			ColumnIds: columnIds,
		}

		_, err := m.boardCol.InsertOne(sc, newBoard)
		if err != nil {
			return fmt.Errorf("could not insert board: %v", err)
		}

		return nil
	})

	return err
}

func (m *MongoDb) EditBoard(ctx context.Context, board *store.Board) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) DeleteBoard(ctx context.Context, boardId string) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) GetBoard(ctx context.Context, boardName string) (*store.Board, error) {

	pipeline := mongo.Pipeline{
		// Match board by name
		{{"$match", bson.D{{"name", boardName}}}},

		// Lookup columns (joining by column ObjectIDs in the board document)
		{{"$lookup", bson.D{
			{"from", "columns"},       // Lookup from the columns collection
			{"localField", "columns"}, // The field from the board (list of ObjectIDs)
			{"foreignField", "_id"},   // The field in the columns collection
			{"as", "columns_info"},    // Output the result as "columns_info"
		}}},

		// Lookup cards (joining by column ObjectIDs)
		{{"$lookup", bson.D{
			{"from", "cards"},                  // Lookup from the cards collection
			{"localField", "columns_info._id"}, // The field from columns_info
			{"foreignField", "columnId"},       // The field in the cards collection
			{"as", "columns_info.cards"},       // Output the result as "cards" inside columns
		}}},
	}

	// Run the aggregation query
	cursor, err := m.boardCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error running aggregation: %v", err)
	}
	defer cursor.Close(ctx)

	// Process the result
	var board Board
	if cursor.Next(ctx) {
		if err := cursor.Decode(&board); err != nil {
			return nil, fmt.Errorf("error decoding result: %v", err)
		}
	} else {
		// If no documents are found, return an error
		return nil, &store.NoQueryResultsError{}
	}

	retBoard := &store.Board{
		Name: board.Name,
	}

	for _, column := range board.Columns {
		cards := make([]*store.Card, len(column.Cards))
		for _, card := range column.Cards {
			cards = append(cards, &store.Card{
				Id:          card.Id.Hex(),
				Title:       card.Title,
				Description: card.Description,
				Index:       card.Index,
			})
		}

		retBoard.Columns = append(retBoard.Columns, &store.Column{
			Id:    column.Id.Hex(),
			Name:  column.Name,
			Index: column.Index,
			Cards: cards,
		})
	}

	return retBoard, errors.New(`Not implemented`)
}
