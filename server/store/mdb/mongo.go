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

func (m *MongoDb) AddCard(ctx context.Context, boardName, columnIdStr string, cardTitle string) (*store.Card, error) {
	var board Board
	err := m.boardCol.FindOne(ctx, bson.M{"name": boardName}).Decode(&board)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.NewNotFoundError("board", boardName)
		}
		return nil, fmt.Errorf("failed to find board: %v", err)
	}

	columnId, err := primitive.ObjectIDFromHex(columnIdStr)
	if err != nil {
		return nil, store.NewBadRequestError(fmt.Sprintf("invalid column id: %s", columnIdStr))
	}

	found := false
	for _, foundId := range board.ColumnIds {
		if foundId == columnId {
			found = true
			break
		}
	}

	if !found {
		return nil, store.NewBadRequestError(fmt.Sprintf("board does not have column: ", columnIdStr))
	}

	count, err := m.cardCol.CountDocuments(ctx, bson.M{"columnId": columnId})
	if err != nil {
		return nil, fmt.Errorf("failed to count documents in target column: %w", err)
	}

	newCard := &Card{
		ColumnId: columnId,
		Title:    cardTitle,
		Index:    int(count),
	}

	result, err := m.cardCol.InsertOne(ctx, newCard)
	if err != nil {
		return nil, fmt.Errorf("failed to insert card: %v", err)
	}

	return &store.Card{
		Id:    result.InsertedID.(primitive.ObjectID).Hex(),
		Title: cardTitle,
		Index: 0,
	}, nil
}

func (m *MongoDb) EditCard(ctx context.Context, boardId, columnId, card *store.Card) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) MoveCard(ctx context.Context, boardName, toColumnIdStr, cardIdStr string, newIndex int) error {
	// Verify inputs are good
	var board Board
	err := m.boardCol.FindOne(ctx, bson.M{"name": boardName}).Decode(&board)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return store.NewNotFoundError("board", boardName)
		}
		return fmt.Errorf("error finding board by name %s: %w", boardName, err)
	}

	cardId, err := primitive.ObjectIDFromHex(cardIdStr)
	if err != nil {
		return store.NewBadRequestError("invalid card id")
	}
	var card Card
	err = m.cardCol.FindOne(ctx, bson.M{"_id": cardId}).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return store.NewNotFoundError("card", cardIdStr)
		}
		return fmt.Errorf("error finding board by id %s: %w", cardId, err)
	}
	if !contains(board.ColumnIds, card.ColumnId) {
		return store.NewBadRequestError(fmt.Sprintf("board does not have card: %s", cardIdStr))
	}

	toColumnId, err := primitive.ObjectIDFromHex(toColumnIdStr)
	if err != nil {
		return store.NewBadRequestError("invalid to column id")
	}
	if !contains(board.ColumnIds, toColumnId) {
		return store.NewBadRequestError(fmt.Sprintf("board does not have column: %s", toColumnId))
	}

	count, err := m.cardCol.CountDocuments(ctx, bson.M{"columnId": toColumnId})
	if err != nil {
		return fmt.Errorf("failed to count documents in target column: %w", err)
	}

	if newIndex < 0 || newIndex > int(count) {
		newIndex = int(count)
	}

	if card.ColumnId == toColumnId {
		if newIndex < card.Index {
			_, err = m.cardCol.UpdateMany(
				ctx,
				bson.M{
					"columnId": card.ColumnId,
					"index":    bson.M{"$gte": newIndex, "$lt": card.Index},
				},
				bson.M{
					"$inc": bson.M{"index": 1},
				},
			)
		} else if newIndex > card.Index {
			_, err = m.cardCol.UpdateMany(
				ctx,
				bson.M{
					"columnId": card.ColumnId,
					"index":    bson.M{"$lte": newIndex, "$gt": card.Index},
				},
				bson.M{
					"$inc": bson.M{"index": -1},
				},
			)
		}
		if err != nil {
			return fmt.Errorf("error shifting card indices: %w", err)
		}
	} else { // Moving from one column to another
		_, err := m.cardCol.UpdateMany(
			ctx,
			bson.M{
				"columnId": card.ColumnId,
				"index":    bson.M{"$gt": card.Index},
			},
			bson.M{
				"$inc": bson.M{"index": -1},
			},
		)
		if err != nil {
			return fmt.Errorf("error shifting card indices in from column: %w", err)
		}

		_, err = m.cardCol.UpdateMany(
			ctx,
			bson.M{
				"columnId": toColumnId,
				"index":    bson.M{"$gte": newIndex},
			},
			bson.M{
				"$inc": bson.M{"index": 1},
			},
		)
		if err != nil {
			return fmt.Errorf("error shifting card indices in to column: %w", err)
		}
	}
	_, err = m.cardCol.UpdateOne(
		ctx,
		bson.M{"_id": cardId},
		bson.M{
			"$set": bson.M{
				"columnId": toColumnId,
				"index":    newIndex,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update card index: %w", err)
	}

	return err
}

func contains(haystack []primitive.ObjectID, needle primitive.ObjectID) bool {
	for _, id := range haystack {
		if id == needle {
			return true
		}
	}
	return false
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

func (m *MongoDb) GetColumns(ctx context.Context, boardName string) ([]*store.Column, error) {
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
			"$addFields": bson.M{
				"columns.cards": bson.M{
					"$sortArray": bson.M{
						"input": "$columns.cards", // The array to sort
						"sortBy": bson.M{
							"index": 1, // Sort by the index field in ascending order
						},
					},
				},
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
		return nil, store.NewNotFoundError("board", name)
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
