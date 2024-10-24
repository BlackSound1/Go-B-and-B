package dbrepo

import (
	"context"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/models"
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
