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

func (m *MongoDb) GetCardCount(ctx context.Context, columnIdStr string) (int, error) {
	columnId, err := primitive.ObjectIDFromHex(columnIdStr)
	if err != nil {
		return 0, store.NewBadRequestError(fmt.Sprintf("invalid column id: %s", columnIdStr))
	}

	count, err := m.cardCol.CountDocuments(ctx, bson.M{"columnId": columnId})
	if err != nil {
		return 0, fmt.Errorf("failed to count documents in target column: %w", err)
	}
	return int(count), nil
}

func (m *MongoDb) AddCard(ctx context.Context, columnIdStr string, cardTitle string) (*store.Card, error) {
	columnId, err := primitive.ObjectIDFromHex(columnIdStr)
	if err != nil {
		return nil, store.NewBadRequestError(fmt.Sprintf("invalid column id: %s", columnIdStr))
	}

	count, err := m.GetCardCount(ctx, columnIdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents in target column: %w", err)
	}

	newCard := &card{
		ColumnId: columnId,
		Title:    cardTitle,
		Index:    int(count),
	}

	result, err := m.cardCol.InsertOne(ctx, newCard)
	if err != nil {
		return nil, fmt.Errorf("failed to insert card: %w", err)
	}

	return &store.Card{
		Id:    result.InsertedID.(primitive.ObjectID).Hex(),
		Title: cardTitle,
		Index: 0,
	}, nil
}

func (m *MongoDb) EditCard(ctx context.Context, card *store.Card) error {
	updateFields := bson.M{
		"title":       card.Title,
		"description": card.Description,
	}

	cardId, err := primitive.ObjectIDFromHex(card.Id)

	updateResult, err := m.cardCol.UpdateOne(
		ctx,
		bson.M{"_id": cardId},        // Filter to find the specific card
		bson.M{"$set": updateFields}, // $set operator to update specific fields
	)
	if err != nil {
		return fmt.Errorf(`Unable to update card: %w`, err)
	}
	if updateResult.MatchedCount == 0 {
		return store.NewNotFoundError("card", card.Id)
	}

	return nil
}

func (m *MongoDb) MoveCard(ctx context.Context, toColumnIdStr, cardIdStr string, newIndex int) error {
	cardId, err := primitive.ObjectIDFromHex(cardIdStr)
	if err != nil {
		return store.NewBadRequestError("invalid card id")
	}

	var card card
	err = m.cardCol.FindOne(ctx, bson.M{"_id": cardId}).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return store.NewNotFoundError("card", cardIdStr)
		}
		return fmt.Errorf("error finding card by id %s: %w", cardId, err)
	}

	toColumnId, err := primitive.ObjectIDFromHex(toColumnIdStr)
	if err != nil {
		return store.NewBadRequestError("invalid to column id")
	}

	count, err := m.GetCardCount(ctx, toColumnIdStr)
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

func (m *MongoDb) DeleteCard(ctx context.Context, boardName, columnIdStr, cardIdStr string) error {
	return errors.New(`Not implemented`)
}

func (m *MongoDb) GetCard(ctx context.Context, cardIdStr string) (*store.Card, error) {
	cardId, err := primitive.ObjectIDFromHex(cardIdStr)
	if err != nil {
		return nil, store.NewBadRequestError(fmt.Sprintf("invalid card id: %s", cardIdStr))
	}

	var card card
	err = m.cardCol.FindOne(ctx, bson.M{"_id": cardId}).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.NewNotFoundError("card", cardIdStr)
		} else {
			return nil, fmt.Errorf("unexpected error getting card from storage: %w", err)
		}
	}

	return &store.Card{
		Id:          cardIdStr,
		Title:       card.Title,
		Description: card.Description,
		Index:       card.Index,
	}, nil
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

func (m *MongoDb) GetColumn(ctx context.Context, columnIdStr string) (*store.Column, error) {
	columnId, err := primitive.ObjectIDFromHex(columnIdStr)
	if err != nil {
		return nil, store.NewBadRequestError(fmt.Sprintf("invalid column id: %s", columnIdStr))
	}

	var column column
	err = m.columnCol.FindOne(ctx, bson.M{"_id": columnId}).Decode(&column)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.NewNotFoundError("column", columnIdStr)
		}
		return nil, fmt.Errorf("unexpected error getting board: %w", err)
	}

	return &store.Column{
		Id:    columnIdStr,
		Name:  column.Name,
		Index: column.Index,
	}, nil
}

func (m *MongoDb) GetColumns(ctx context.Context, boardName string) ([]*store.Column, error) {
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"name": boardName}}},
		{{"$lookup", bson.M{
			"from":         "columns",
			"localField":   "columnIds",
			"foreignField": "_id",
			"as":           "columns",
		}}},
	}

	cursor, err := m.boardCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("Aggregation error getting board with columns: %w", err)
	}
	defer cursor.Close(ctx)

	var board board
	if cursor.Next(ctx) {
		err := cursor.Decode(&board)
		if err != nil {
			return nil, fmt.Errorf("unexpected error decoding board with columns: %w", err)
		}
	}

	columns := make([]*store.Column, 0, len(board.Columns))
	for _, column := range board.Columns {
		columns = append(columns, &store.Column{
			Id:    column.Id.Hex(),
			Name:  column.Name,
			Index: column.Index,
		})
	}
	return columns, nil
}

func (m *MongoDb) AddBoard(ctx context.Context, boardDTO *store.Board) error {
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("could not start session: %v", err)
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		var columnIds []primitive.ObjectID
		for _, col := range boardDTO.Columns {
			newColumn := &column{
				Name:  col.Name,
				Index: col.Index,
			}
			colRes, err := m.columnCol.InsertOne(sc, newColumn)
			if err != nil {
				return fmt.Errorf("could not insert column: %v", err)
			}
			colId := colRes.InsertedID.(primitive.ObjectID)
			col.Id = colId.Hex()
			columnIds = append(columnIds, colId)

			for _, c := range col.Cards {
				newCard := &card{
					Title:       c.Title,
					Description: c.Description,
					Index:       c.Index,
					ColumnId:    colId,
				}
				cardRes, err := m.cardCol.InsertOne(sc, newCard)
				if err != nil {
					return fmt.Errorf("could not insert card: %v", err)
				}
				c.Id = cardRes.InsertedID.(primitive.ObjectID).Hex()
			}
		}

		newBoard := &board{
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

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"name": name,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "columns",
				"localField":   "columnIds",
				"foreignField": "_id",
				"as":           "columns",
			},
		},
		{
			"$unwind": "$columns",
		},
		{
			"$lookup": bson.M{
				"from":         "cards",
				"localField":   "columns._id",
				"foreignField": "columnId",
				"as":           "columns.cards",
			},
		},
		{
			"$addFields": bson.M{
				"columns.cards": bson.M{
					"$sortArray": bson.M{
						"input": "$columns.cards",
						"sortBy": bson.M{
							"index": 1,
						},
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id":       "$_id",
				"name":      bson.M{"$first": "$name"},
				"columnIds": bson.M{"$first": "$columnIds"},
				"columns":   bson.M{"$push": "$columns"},
			},
		},
	}

	cursor, err := m.boardCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	// Check if any board was found
	var result board
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
