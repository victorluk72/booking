package models

import (
	"time"
)

// User is the model for users
type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	AccessLevel int
}

// Room is the model for room
type Room struct {
	ID        int
	RoomName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Restriction is the model for restriction
type Restriction struct {
	ID              int
	RestrictionName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Reservation is the model for reservation
type Reservation struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	RoomID    int
	Room      Room
	CreatedAt time.Time
	UpdatedAt time.Time
	Processed int
}

// RoomRestriction is the model for room restriction
type RoomRestriction struct {
	ID            int
	StartDate     time.Time
	EndDate       time.Time
	ReservationID int
	Reservation   Reservation
	RestrictionID int
	Restriction   Restriction
	RoomID        int
	Room          Room
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// MailData contains all data related to email message
type MailData struct {
	To      string
	From    string
	Subject string
	Content string
}
