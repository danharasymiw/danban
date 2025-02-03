package mdb

import "go.mongodb.org/mongo-driver/bson/primitive"

type board struct {
	Id        primitive.ObjectID   `bson:"_id,omitempty"`
	Name      string               `bson:"name"`
	ColumnIds []primitive.ObjectID `bson:"columnIds,omitempty"`
	Columns   []column             `bson:"columns,omitempty"` // This is just here for the aggregation, never stored
}

type column struct {
	Id    primitive.ObjectID `bson:"_id,omitempty"`
	Index int                `bson:"index"`
	Name  string             `bson:"name"`
	Cards []card             `bson:"cards,omitempty"` // This is just here for the aggregation, never stored
}

type card struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	Index       int                `bson:"index"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	ColumnId    primitive.ObjectID `bson:"columnId"`
}
