package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role represents user roles for RBAC
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// UserStatus represents account status
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusBanned   UserStatus = "banned"
)

// User is the MongoDB document model
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"   json:"id"`
	FirstName string             `bson:"first_name"      json:"first_name"`
	LastName  string             `bson:"last_name"       json:"last_name"`
	Email     string             `bson:"email"           json:"email"`
	Password  string             `bson:"password"        json:"-"` // never serialise
	Phone     string             `bson:"phone,omitempty" json:"phone,omitempty"`
	Role      Role               `bson:"role"            json:"role"`
	Status    UserStatus         `bson:"status"          json:"status"`
	AvatarURL string             `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	CreatedAt time.Time          `bson:"created_at"      json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"      json:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"` // soft-delete
}

// CollectionName returns the MongoDB collection name
func (User) CollectionName() string {
	return "users"
}
