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

type Storage struct {
	db *sql.DB
}

var emptyID = 0

// New creates a new Storage instance
func New(storagePath string) (*Storage, error) {
	const op = "storage.postgresql.New"

	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

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
	birthDate time.Time,
	description string,
	photo []byte,
) (int, error) {
	const op = "storage.postgresql.CreateUser"

	var schema string
	switch institute {
	case "grafit":
		schema = "grafit"
	case "giredmet":
		schema = "giredmet"
	default:
		return emptyID, fmt.Errorf("%s: unknown institute %s", op, institute)
	}

	var id int

	query := fmt.Sprintf(
		`INSERT INTO %s.workers (
		surname, name, middle_name,
		email, phone_number, cabinet,
		position, department,
		birth_date, description, photo)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
		schema,
	)

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

	var schema string
	switch institute {
	case "grafit":
		schema = "grafit"
	case "giredmet":
		schema = "giredmet"
	default:
		return fmt.Errorf("%s: unknown institute %s", op, institute)
	}

	query := fmt.Sprintf(
		"DELETE FROM %s.workers WHERE email = $1",
		schema,
	)

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
	birthDate time.Time,
	description string,
	photo []byte,
) error {
	const op = "storage.postgresql.UpdateUser"

	var schema string
	switch institute {
	case "grafit":
		schema = "grafit"
	case "giredmet":
		schema = "giredmet"
	default:
		return fmt.Errorf("%s: unknown institute %s", op, institute)
	}

	query := fmt.Sprintf(
		`UPDATE %s.workers SET 
			surname = $1, 
			name = $2, 
			middle_name = $3, 
			email = $4, 
			phone_number = $5, 
			cabinet = $6, 
			position = $7, 
			department = $8,
			birth_date = $9,
			description = $10,
			photo = $11
		WHERE email = $12`,
		schema,
	)

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

	var schema string
	switch institute {
	case "grafit":
		schema = "grafit"
	case "giredmet":
		schema = "giredmet"
	default:
		return models.EmptyUser, fmt.Errorf("%s: unknown institute %s", op, institute)
	}

	query := fmt.Sprintf(
		`SELECT id, surname, name, middle_name, email, phone_number, cabinet, position, department, birth_date, description, photo
		FROM %s.workers WHERE email = $1`,
		schema,
	)

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

func (s *Storage) GetAllUsers(ctx context.Context, institute string, department string) ([]models.User, error) {
	const op = "storage.postgresql.GetAllUsers"

	var schema string
	switch institute {
	case "grafit":
		schema = "grafit"
	case "giredmet":
		schema = "giredmet"
	default:
		return nil, fmt.Errorf("%s: unknown institute %s", op, institute)
	}

	var query string
	var args []interface{}

	if department == "" {
		// Если отдел не указан, возвращаем всех пользователей
		query = fmt.Sprintf(
			`SELECT surname, name, middle_name, email, phone_number, cabinet, position, department
			FROM %s.workers ORDER BY surname, name`,
			schema,
		)
	} else {
		// Если отдел указан, фильтруем по нему
		query = fmt.Sprintf(
			`SELECT id, surname, name, middle_name, email, phone_number, cabinet, position, department
			FROM %s.workers WHERE department = $1 ORDER BY surname, name`,
			schema,
		)
		args = append(args, department)
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
