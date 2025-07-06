package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"telephone-book/internal/storage"

	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

const emptyID = 0

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
		"INSERT INTO %s.workers (surname, name, middle_name, email, phone_number, cabinet, position, department) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
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
	).Scan(&id)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return emptyID, storage.ErrUserAlreadyExists
		}
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
