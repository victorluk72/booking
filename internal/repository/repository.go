package repository

import (
	"time"

	"github.com/victorluk72/booking/internal/models"
)

//Interface for database repo. It should cover all possible functions for DB CRUD
type DatabaseRepo interface {
	AllUsers() bool

	InsertReservstion(res models.Reservation) (int, error)
	InsertRoomRestriction(r models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomByID(room_id int) (models.Room, error)
	GetAllRooms() ([]models.Room, error)
	GetUserByID(id int) (models.User, error)
	UpdateUser(u models.User) error
	Authenticate(email, testPassword string) (int, string, error)

	AllReservations() ([]models.Reservation, error)
	NewReservations() ([]models.Reservation, error)
	GetReservationByID(id int) (models.Reservation, error)
	UpdateReservation(u models.Reservation) error
	DeleteReservation(id int) error
	UpdateProcessedForReservation(id, processed int) error

	GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error)
}
