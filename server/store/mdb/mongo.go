package mdb

import (
	"context"
	"errors"
	"fmt"
	"os"

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
	uri := "mongodb://localhost:27017"
	deployedMongoUrl := os.Getenv("MONGO_URL")
	if deployedMongoUrl != `` {
		uri = deployedMongoUrl
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
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

func (m *MongoDb) AddCard(ctx context.Context, boardName, columnIdStr string, card *store.Card) error {
	var board Board
	err := m.boardCol.FindOne(ctx, bson.M{"name": boardName}).Decode(&board)
	if err != nil {
		return fmt.Errorf("failed to find board: %v", err)
	}

	columnId, err := primitive.ObjectIDFromHex(columnIdStr)
	if err != nil {
		return fmt.Errorf("invalid column id: %s", columnIdStr)
	}

	found := false
	for _, foundId := range board.ColumnIds {
		if foundId == columnId {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("column with id %s not found in board", columnIdStr)
	}

	newCard := &Card{
		ColumnId: columnId,
		Title:    card.Title,
	}

	result, err := m.cardCol.InsertOne(ctx, newCard)
	if err != nil {
		return fmt.Errorf("failed to insert card: %v", err)
	}

	// TODO - don't like this... return a whole new card instead of modifying the state of given card
	card.Id = result.InsertedID.(primitive.ObjectID).Hex()

	return nil
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
			newColumn := &Column{
				Name:  column.Name,
				Index: column.Index,
			}
			colRes, err := m.columnCol.InsertOne(sc, newColumn)
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

func (m *MongoDb) GetBoard(ctx context.Context, name string) (*store.Board, error) {

	// Set up the aggregation pipeline
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"name": name,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "columns",   // Assuming columns are in a separate collection
				"localField":   "columnIds", // Board has a list of ColumnIds
				"foreignField": "_id",       // Column has an _id field
				"as":           "columns",   // The result will be stored in a "columns" field
			},
		},
		{
			"$unwind": "$columns", // If you want to unwind the columns to handle them as separate documents
		},
		{
			"$lookup": bson.M{
				"from":         "cards",         // Assuming cards are in a separate collection
				"localField":   "columns._id",   // The column Id in the column document
				"foreignField": "columnId",      // Card has a columnId that references the column
				"as":           "columns.cards", // The result will be stored in the "cards" field in each column
			},
		},
		{
			"$group": bson.M{
				"_id":       "$_id",                    // Group by the board's ID
				"name":      bson.M{"$first": "$name"}, // Take the first occurrence of the board's name
				"columnIds": bson.M{"$first": "$columnIds"},
				"columns":   bson.M{"$push": "$columns"}, // Combine the columns back into an array
			},
		},
	}

	cursor, err := m.boardCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	// Check if any board was found
	var result Board
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode result: %w", err)
		}
	}

	if result.Id.IsZero() {
		return nil, &store.NoQueryResultsError{}
	}

	columns := make([]*store.Column, 0, len(result.Columns))
	for _, column := range result.Columns {
		cards := make([]*store.Card, 0, len(column.Cards))
		for _, card := range column.Cards {
			cards = append(cards, &store.Card{
				Id:          card.Id.Hex(),
				Title:       card.Title,
				Description: card.Description,
				Index:       card.Index,
			})
		}
		columns = append(columns, &store.Column{
			Id:    column.Id.Hex(),
			Name:  column.Name,
			Cards: cards,
		})
	}

	return &store.Board{
		Name:    result.Name,
		Columns: columns,
	}, nil
}
