package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a new reservation record into the database.
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// Allows for a 3 second timeout of the query. Needs to be able to cancel
	// the query if it takes too long or else the connection might have been lost
	// and data can be corrupted
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	var newID int

	stmt := `
		INSERT INTO 
			reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id
	`

	// Instead of Exec(), use QueryRowContext() to allow for the 3 second timeout.
	// Also, it allows us to return the id of the reservation
	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// InsertRoomRestriction inserts a new room restriction record into the database.
func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO 
			room_restrictions (start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := m.DB.ExecContext(
		ctx,
		stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		r.RestrictionID,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDates takes in a start and end date and checks to see if there
// is any availability in the room_restrictions table for that date range for a given room ID.
// If there are no rows, it means there is availability.
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		SELECT
			COUNT(id)
		FROM
			room_restrictions
		WHERE
			room_id = $1 AND
			$2 < end_date AND $3 > start_date
	`

	// Run query
	row := m.DB.QueryRowContext(
		ctx,
		stmt,
		roomID,
		start,
		end,
	)

	var numRows int

	err := row.Scan(&numRows) // Get the count of rows
	if err != nil {
		return false, err
	}

	// If there are no rows, then there is availbility
	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of room if any are available
// for the given start and end dates. If there are no rows, it means there are no
// available rooms for the given date range.
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	stmt := `
		SELECT 
			R.id, R.room_name 
		FROM 
			rooms R 
		WHERE 
			R.ID NOT IN 
				(SELECT RR.room_id FROM room_restrictions RR WHERE $1 < RR.end_date AND $2 > RR.start_date)
	`

	// Run query
	rows, err := m.DB.QueryContext(
		ctx,
		stmt,
		start,
		end,
	)
	if err != nil {
		return rooms, err
	}

	// Loop through rows
	for rows.Next() {
		var room models.Room

		// Get the ID and room name from the current room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}

		// Add the room to the slice
		rooms = append(rooms, room)
	}

	// Check for errors one last time at end
	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRoomByID retrieves a room record from the database by its ID.
func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	stmt := `
		SELECT
			id, room_name, created_at, updated_at
		FROM
			rooms
		WHERE
			id = $1
	`

	// Run query
	row := m.DB.QueryRowContext(ctx, stmt, id)

	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err != nil {
		return room, err
	}

	return room, nil
}

// GetUserByID retrieves a user record from the database by their ID.
func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			id, first_name, last_name, email, password, access_level, created_at, updated_at
		FROM
			users
		WHERE
			id = $1	
	`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User

	// Try to scan the row into the user
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}

	return u, nil
}

// UpdateUser updates a user record in the database.
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE
			users
		SET
			first_name = $1,
			last_name = $2,
			email = $3,
			access_level = $4,
			updated_at = $5
		WHERE
			id = $6
	`

	// Try to update the user
	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(),
		u.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// Authenticate verifies user credentials by checking the email and password.
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int                // Hold ID of authenticated user
	var hashedPassword string // Hold hashed password of authenticated user

	// Get user from database
	row := m.DB.QueryRowContext(ctx, "SELECT id, password FROM users WHERE email = $1", email)

	// Try to scan the row data
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	// Check password against the hashed password in the database
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))

	if err == bcrypt.ErrMismatchedHashAndPassword {
		// If the passwords don't match
		return 0, "", errors.New("incorrect password")

	} else if err != nil {
		// If there is another error
		return 0, "", err
	}

	// If no error, user is authenticated
	return id, hashedPassword, nil
}

// AllReservations retrieves all reservations from the database.
func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		SELECT
			r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date,
		 	r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		FROM 
			reservations r
		JOIN 
			rooms rm 
				ON (r.room_id = rm.id)
		ORDER BY 
			r.start_date ASC
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	// For each of the Reservations we got, add each to the reservations slice
	for rows.Next() {
		var item models.Reservation

		err := rows.Scan(
			&item.ID,
			&item.FirstName,
			&item.LastName,
			&item.Email,
			&item.Phone,
			&item.StartDate,
			&item.EndDate,
			&item.RoomID,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Processed,
			&item.Room.ID,
			&item.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, item)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// AllNewReservations retrieves all new, unprocessed reservations from the database.
func (m *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		SELECT
			r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date,
		 	r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		FROM 
			reservations r
		LEFT JOIN 
			rooms rm 
				ON (r.room_id = rm.id)
		WHERE
			r.processed = 0
		ORDER BY 
			r.start_date ASC
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	// For each of the Reservations we got, add each to the reservations slice
	for rows.Next() {
		var item models.Reservation

		err := rows.Scan(
			&item.ID,
			&item.FirstName,
			&item.LastName,
			&item.Email,
			&item.Phone,
			&item.StartDate,
			&item.EndDate,
			&item.RoomID,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Processed,
			&item.Room.ID,
			&item.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, item)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// GetReservationByID retrieves a reservation record from the database by ID.
func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `
		SELECT 
			r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date,
			r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		FROM 
			reservations r
		LEFT JOIN 
			rooms rm 
				ON (r.room_id = rm.id)
		WHERE
			r.id = $1
	`

	row := m.DB.QueryRowContext(ctx, query, id)

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

// UpdateReservation updates a reservation record in the database.
func (m *postgresDBRepo) UpdateReservation(r models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE
			reservations
		SET
			first_name = $1,
			last_name = $2,
			email = $3,
			phone = $4,
			updated_at = $5
		WHERE
			id = $6
	`

	_, err := m.DB.ExecContext(ctx, query,
		r.FirstName,
		r.LastName,
		r.Email,
		r.Phone,
		time.Now(),
		r.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// DeleteReservation deletes a reservation record from the database by ID.
func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM reservations WHERE id = $1`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProcessedForReservation updates the processed status of a reservation
// record in the database by ID.
func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE 
			reservations
		SET 
			processed = $1
		WHERE 
			id = $2
	`

	_, err := m.DB.ExecContext(ctx, query, processed, id)
	if err != nil {
		return err
	}

	return nil
}

// AllRooms retrieves all rooms from the database, ordered by room name.
func (m *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
		SELECT
			id, room_name, created_at, updated_at
		FROM
			rooms
		ORDER BY
			room_name
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}
	defer rows.Close()

	for rows.Next() {
		var rm models.Room
		err := rows.Scan(
			&rm.ID,
			&rm.RoomName,
			&rm.CreatedAt,
			&rm.UpdatedAt,
		)

		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, rm)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRestrictionsForRoomByDate retrieves a list of room restrictions for a given room ID
// within a specified date range.
func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// coalesce lets us return 0 if there is no reservation_id
	query := `
		SELECT
			id, COALESCE(reservation_id, 0), restriction_id, room_id, start_date, end_date
		FROM
			room_restrictions
		WHERE
			$1 < end_date AND $2 >= start_date
		AND
			room_id = $3
	`

	var restrictions []models.RoomRestriction

	rows, err := m.DB.QueryContext(ctx, query, start, end, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction
		err := rows.Scan(
			&r.ID,
			&r.ReservationID, // Can be nil, so we use coalesce
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)
		if err != nil {
			return nil, err
		}

		restrictions = append(restrictions, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return restrictions, nil
}
