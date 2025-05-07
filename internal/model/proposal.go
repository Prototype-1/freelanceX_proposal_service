package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Proposal struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	ClientID    string             `bson:"client_id"`
	FreelancerID string           `bson:"freelancer_id"`
	TemplateID  primitive.ObjectID `bson:"template_id"`
	Title       string             `bson:"title"`
	Content     string             `bson:"content"`
	Status      string             `bson:"status"` // "draft" | "sent" | "accepted" | "rejected"
	Version     int                `bson:"version"`
	Deadline     time.Time          `bson:"deadline"`
	Sections  []Section  `bson:"sections,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

type Template struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	OwnerID   string             `bson:"owner_id"`
	Title     string             `bson:"title"`
	Sections  []Section          `bson:"sections"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type Section struct {
	Heading string `bson:"heading"`
	Body    string `bson:"body"`
}
