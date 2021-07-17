package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/victorluk72/booking/internal/models"
	"golang.org/x/crypto/bcrypt"
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
	defer rows.Close()

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

func (m *postgresDBRepo) GetAllRooms() ([]models.Room, error) {
	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//variable for rooms - slise of models (from model Room)
	var rooms []models.Room

	query := `select id, room_name, created_at, updated_at from rooms order by room_name`

	//get rows with list of rooms
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}

	defer rows.Close()

	//Scan all rows and asign to slice of rooms
	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID,
			&room.RoomName,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
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

// UpdateUser updates user in the database
func (m *postgresDBRepo) UpdateUser(u models.User) error {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update users set first_name=$1 last_name=$2, email = $3, access_level = $4, updated_at = $5`

	_, err := m.DB.ExecContext(ctx, query, u.FirstName, u.LastName, u.Email, u.AccessLevel, time.Now())
	if err != nil {
		return err
	}

	return nil
}

// Authenticate check if password and email at matching
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//variable for authenticated user and it's password
	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)

	//Scan into variables
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	//Compare entered password (hashed) to password in DB (used build in function )
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))

	//logic if password mismatched
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err

	}

	//logic if password matched
	return id, hashedPassword, nil

}

// AllReservations returns the slice of all reservations
func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//variable for all reservations from db
	var reservations []models.Reservation

	query := `select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
	          r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
			  rm.id, rm.room_name
			  from reservations r
			  left join rooms rm on (r.room_id = rm.id)
			  order by r.start_date asc  
	         `
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	//scan our rows and put fields into variable
	for rows.Next() {

		var i models.Reservation

		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}

		//happy path  - apent all rows
		reservations = append(reservations, i)
	}

	//do another error check
	if err = rows.Err(); err != nil {
		return reservations, err
	}

	//return list of reservations and error
	return reservations, nil
}

// NewReservations returns the slice of new reservations
func (m *postgresDBRepo) NewReservations() ([]models.Reservation, error) {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//variable for all reservations from db
	var reservations []models.Reservation

	query := `select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
	          r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
			  rm.id, rm.room_name
			  from reservations r
			  left join rooms rm on (r.room_id = rm.id)
			  where r.processed=0
			  order by r.start_date asc  
	         `
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	//scan our rows and put fields into variable
	for rows.Next() {

		var i models.Reservation

		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}

		//happy path  - apent all rows
		reservations = append(reservations, i)
	}

	//do another error check
	if err = rows.Err(); err != nil {
		return reservations, err
	}

	//return list of reservations and error
	return reservations, nil
}

// GetReservationByID returns single reservation (model) by ID
func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//variable to hold informaton about single reservation
	var res models.Reservation

	query := `select r.id, r.first_name, r.last_name, r.email, r.phone, 
	          r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at,
			  r.processed, rm.id, rm.room_name
			  from reservations r
			  left join rooms rm on (r.room_id = rm.id)
			  where r.id=$1`

	row := m.DB.QueryRowContext(ctx, query, id)

	//Scan into variables
	err := row.Scan(
		&res.ID,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.RoomID,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processed,
		&res.Room.ID,
		&res.Room.RoomName,
	)
	if err != nil {
		return res, err
	}

	return res, nil

}

// UpdateReservation updates model for reservation in the database
func (m *postgresDBRepo) UpdateReservation(r models.Reservation) error {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update reservations set first_name=$1, last_name=$2, email = $3, 
	          phone = $4, updated_at = $5 where id = $6 `

	_, err := m.DB.ExecContext(ctx, query, r.FirstName, r.LastName, r.Email, r.Phone, time.Now(), r.ID)
	if err != nil {
		return err
	}

	return nil
}

// DeleteReservation deletes reservation by from the database
func (m *postgresDBRepo) DeleteReservation(id int) error {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from reservations where id=$1`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProcessedForReservation updates field process for single reservation
func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update reservations set processed=$1 where id=$2`

	_, err := m.DB.ExecContext(ctx, query, processed, id)
	if err != nil {
		return err
	}

	return nil
}

// GetRestrictionsForRoomByDate return current restriction for date range
func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {

	//If transaction takes longeer than 3 seconds cancel it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	query := `select id, coalesce(reservation_id, 0), restriction_id, room_id, start_date, end_date
	          from room_restrictions where $1 < end_date and $2 >= start_date 
			  and room_id = $3`

	rows, err := m.DB.QueryContext(ctx, query, start, end, roomID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	//Scan to variable
	for rows.Next() {
		var r models.RoomRestriction
		err := rows.Scan(
			&r.ID,
			&r.ReservationID,
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)

		if err != nil {
			return nil, err
		}

		//Build my restrictions data
		restrictions = append(restrictions, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return restrictions, nil

}
