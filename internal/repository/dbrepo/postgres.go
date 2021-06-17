package dbrepo

import (
	"context"
	"time"

	"github.com/victorluk72/booking/internal/models"
)

//-- we define all functions for my interface DatabaseRepo (see file dbripo.go)

//AllUsers
func (m *postgresDBRepo) AllUsers() bool {
	return true

}

// InsertReservstion inserts reservation details into database
// This to be executed from corresponded handler (PostReservation)
func (m *postgresDBRepo) InsertReservstion(res models.Reservation) (int, error) {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int

	// Insert into DB statement
	stmt := `insert into reservations (first_name, last_name, email, phone, start_date, end_date, 
		     room_id, created_at, updated_at) 
	         values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now()).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// InsertRoomRestriction inserts room restriction details into database
// This happend immediately after room is reserved
// This to be executed from corresponded handler (PostReservation)
func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Insert into DB statement
	stmt := `insert into room_restrictions (start_date, end_date, room_id, restriction_id,
		     reservation_id, created_at, updated_at) 
			 values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.RestrictionID,
		r.ReservationID,
		time.Now(),
		time.Now())

	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDates returns true when room avaialble and false when it is booked
// This apply to given roomID only (you need to pass room id)
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select count(id) from room_restrictions
	          where room_id = $1 and $2 < end_date and $3 > start_date`

	var numRows int

	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	//this is the case with no result - room is avaialbe
	if numRows == 0 {
		return true, nil
	}

	//this is the case with 1 result - room is Not avaialbe
	return false, nil

}

// SearchAvailabilityForAllRooms search all avaialble room for period of time and return slice of rooms
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//variable for rooms (from model Room)
	var rooms []models.Room

	query := `select r.id, r.room_name from rooms r 
	          where r.id not in 
			  (select rr.room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date)`

	//get rows with list of rooms
	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	//Scan all rows and asign to slice of rooms
	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return rooms, err
		}

		//append single room to slice of rooms
		rooms = append(rooms, room)
	}

	//Additional error check
	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRoomByID returns one room of type models.Room
func (m *postgresDBRepo) GetRoomByID(room_id int) (models.Room, error) {
	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `select id, room_name from rooms where id = $1`

	row := m.DB.QueryRowContext(ctx, query, room_id)

	//Scan into variables
	err := row.Scan(&room.ID, &room.RoomName)
	if err != nil {
		return room, err
	}

	return room, nil

}

// GetUserByID returns user type models.Uset
func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u models.User

	query := `select id, first_name, last_name, email, pasword, access_level
	          created_at, updated_at from users where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	//Scan into variables
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.AccessLevel, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return u, err
	}

	return u, nil

}

//
func (m *postgresDBRepo) UpdateUser(u models.User) error {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return nil
}
