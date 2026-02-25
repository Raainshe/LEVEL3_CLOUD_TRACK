package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServiceLog struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	InstanceName string             `json:"instance_name" bson:"instance_name"`
	Namespace    string             `json:"namespace" bson:"namespace"`
	EventType    string             `json:"event_type" bson:"event_type"`
	FromStatus   string             `json:"from_status,omitempty" bson:"from_status,omitempty"`
	ToStatus     string             `json:"to_status" bson:"to_status"`
	Message      string             `json:"message" bson:"message"`
	Details      string             `json:"details,omitempty" bson:"details,omitempty"`
	Timestamp    time.Time          `json:"timestamp" bson:"timestamp"`
}

type InstanceStatusCache struct {
	InstanceName string    `json:"instance_name" bson:"instance_name"`
	Namespace    string    `json:"namespace" bson:"namespace"`
	Status       string    `json:"status" bson:"status"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}
