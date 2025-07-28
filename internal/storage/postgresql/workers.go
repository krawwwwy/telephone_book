package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"telephone-book/internal/domain/models"
	"telephone-book/internal/storage"
	"time"

	"github.com/lib/pq"
)

func (s *Storage) CreateUser(
	ctx context.Context,
	institute string,
	surname string,
	name string,
	middleName string,
	email string,
	phoneNumber string,
	cabinet string,
	position string,
	department string,
	section string,
	birthDate time.Time,
	description string,
	photo []byte,
) (int, error) {
	const op = "storage.postgresql.CreateUser"

	if err := s.SetSchema(ctx, institute); err != nil {
		return emptyID, err
	}

	var id int

	query := `
		INSERT INTO workers (
			surname, name, middle_name,
			email, phone_number, cabinet,
			position, department, section,
			birth_date, description, photo
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) 
		RETURNING id
		`

	err := s.db.QueryRowContext(
		ctx,
		query,
		surname,
		name,
		middleName,
		email,
		phoneNumber,
		cabinet,
		position,
		department,
		section,
		birthDate,
		description,
		photo,
	).Scan(&id)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return emptyID, storage.ErrUserAlreadyExists
		}
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) DeleteUser(
	ctx context.Context,
	institute string,
	email string,
) error {
	const op = "storage.postgresql.DeleteUser"

	if err := s.SetSchema(ctx, institute); err != nil {
		return err
	}

	query := "DELETE FROM workers WHERE email = $1"

	result, err := s.db.ExecContext(ctx, query, email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	return nil
}

func (s *Storage) UpdateUser(
	ctx context.Context,
	institute string,
	oldEmail string,
	surname string,
	name string,
	middleName string,
	email string,
	phoneNumber string,
	cabinet string,
	position string,
	department string,
	section string,
	birthDate time.Time,
	description string,
	photo []byte,
) error {
	const op = "storage.postgresql.UpdateUser"

	if err := s.SetSchema(ctx, institute); err != nil {
		return err
	}

	query := `UPDATE workers SET 
			surname = $1, 
			name = $2, 
			middle_name = $3, 
			email = $4, 
			phone_number = $5, 
			cabinet = $6, 
			position = $7, 
			department = $8,
			section = $9,
			birth_date = $10,
			description = $11,
			photo = $12
		WHERE email = $13`

	result, err := s.db.ExecContext(
		ctx,
		query,
		surname,
		name,
		middleName,
		email,
		phoneNumber,
		cabinet,
		position,
		department,
		section,
		birthDate,
		description,
		photo,
		oldEmail,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return storage.ErrUserAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	return nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, institute string, email string) (models.User, error) {
	const op = "storage.postgresql.GetUserByEmail"

	if err := s.SetSchema(ctx, institute); err != nil {
		return models.EmptyUser, err
	}

	query := `SELECT id, surname, name, middle_name, email, phone_number, cabinet, position, department, section, birth_date, description, photo
		FROM workers WHERE email = $1`

	var user models.User

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Surname,
		&user.Name,
		&user.MiddleName,
		&user.Email,
		&user.PhoneNumber,
		&user.Cabinet,
		&user.Position,
		&user.Department,
		&user.Section,
		&user.BirthDate,
		&user.Description,
		&user.Photo,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.EmptyUser, storage.ErrUserNotFound
		}
		return models.EmptyUser, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) GetAllUsers(ctx context.Context, institute string, department string, section string) ([]models.User, error) {
	const op = "storage.postgresql.GetAllUsers"

	if err := s.SetSchema(ctx, institute); err != nil {
		return nil, err
	}

	var query string
	var args []interface{}

	if department == "" {
		// Если отдел не указан, возвращаем всех пользователей
		query = `SELECT id, surname, name, middle_name, email, phone_number, cabinet, position, department, section
			FROM workers ORDER BY surname, name`
	} else if section == "" {
		// Если отдел указан, но секция нет - фильтруем по отделу
		query = `SELECT id, surname, name, middle_name, email, phone_number, cabinet, position, department, section
			FROM workers WHERE department = $1 ORDER BY surname, name`
		args = append(args, department)
	} else {
		// Если указаны и отдел, и секция - фильтруем по обоим
		query = `SELECT id, surname, name, middle_name, email, phone_number, cabinet, position, department, section
			FROM workers WHERE department = $1 AND section = $2 ORDER BY surname, name`
		args = append(args, department, section)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.ID,
			&user.Surname,
			&user.Name,
			&user.MiddleName,
			&user.Email,
			&user.PhoneNumber,
			&user.Cabinet,
			&user.Position,
			&user.Department,
			&user.Section,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return users, nil
}
