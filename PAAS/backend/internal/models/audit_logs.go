package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLog struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserEmail     string             `json:"user_email" bson:"user_email"`
	Action        Action             `json:"action" bson:"action"`
	AdminInfo     bool               `json:"admin_info" bson:"admin_info"` // true if the action information is for an admin
	Timestamp     time.Time          `json:"timestamp" bson:"timestamp"`
	RequestMethod string             `json:"request_method,omitempty" bson:"request_method,omitempty"`
	RequestPath   string             `json:"request_path,omitempty" bson:"request_path,omitempty"`
	ClientIP      string             `json:"client_ip,omitempty" bson:"client_ip,omitempty"`
	UserAgent     string             `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
}

type Action struct {
	Action    string `json:"action" bson:"action"`
	Name      string `json:"name" bson:"name"`
	Namespace string `json:"namespace" bson:"namespace"`
	Details   string `json:"details,omitempty" bson:"details,omitempty"`
}
