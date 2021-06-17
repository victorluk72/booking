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
	GetUserByID(id int) (models.User, error)
}
