package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    SenderID   primitive.ObjectID `bson:"sender_id" json:"sender_id"`
    ReceiverID primitive.ObjectID `bson:"receiver_id" json:"receiver_id"`
    Type       string             `bson:"type" json:"type"`
    Text       string             `bson:"text,omitempty" json:"text,omitempty"`
    FileURL    string             `bson:"file_url,omitempty" json:"file_url,omitempty"`
    FileName   string             `bson:"file_name,omitempty" json:"file_name,omitempty"`
    CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}



