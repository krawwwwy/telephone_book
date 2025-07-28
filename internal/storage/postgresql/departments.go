package postgresql

import (
	"context"
	"fmt"
	"telephone-book/internal/domain/models"

	_ "github.com/lib/pq"
)

func (s *Storage) CreateDepartment(ctx context.Context, institute string, name string) (int, error) {
	const op = "storage.postgresql.departments.CreateDepartment"

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

	query := fmt.Sprintf(`INSERT INTO %s.departments (name) VALUES ($1) RETURNING id,`, schema)

	err := s.db.QueryRowContext(ctx, query, name).Scan(&id)

	if err != nil {
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
func (s *Storage) GetAllDepartments(ctx context.Context, institute string) ([]models.Department, error) {
	const op = "storage.postgresql.departments.GetAllDepartments"

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

	query = fmt.Sprintf(`SELECT name FROM %s.departments`, schema)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var departments []models.Department
	for rows.Next() {
		var department models.Department

		err := rows.Scan(&department)

		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}
		departments = append(departments, department)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return departments, nil
}

func (s *Storage) DeleteDepartment(ctx context.Context, institute string, name string) error {
	const op = "storage.postgresql.departments.DeleteDepartment"

	var schema string
	switch institute {
	case "grafit":
		schema = "grafit"
	case "giredmet":
		schema = "giredmet"
	default:
		return fmt.Errorf("%s: unknown institute %s", op, institute)
	}

	query := fmt.Sprintf(`DELETE FROM %s.departments WHERE name = $1`, schema)

	_, err := s.db.ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateDepartment(ctx context.Context, institute string, oldName string, newName string) error {
	const op = "storage.postgresql.departments.UpdateDepartment"

	var schema string
	switch institute {
	case "grafit":
		schema = "grafit"
	case "giredmet":
		schema = "giredmet"
	default:
		return fmt.Errorf("%s: unknown institute %s", op, institute)
	}

	query := fmt.Sprintf(`UPDATE %s.departments SET name = $1 WHERE name = $2`, schema)

	_, err := s.db.ExecContext(ctx, query, newName, oldName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
