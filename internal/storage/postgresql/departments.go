package postgresql

import (
	"context"
	"fmt"
	"telephone-book/internal/domain/models"

	_ "github.com/lib/pq"
)

func (s *Storage) CreateDepartment(ctx context.Context, institute string, name string, sections []string) (int, error) {
	const op = "storage.postgresql.departments.CreateDepartment"

	// Начинаем транзакцию
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return emptyID, fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer tx.Rollback() // Убедимся, что транзакция будет откачена в случае ошибки

	if err := s.SetSchema(ctx, institute); err != nil {
		return emptyID, err
	}

	var id int
	query := `INSERT INTO departments (name) VALUES ($1) RETURNING id`
	err = tx.QueryRowContext(ctx, query, name).Scan(&id)
	if err != nil {
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	query = `INSERT INTO sections (name, parent_id) VALUES ($1, $2)`
	for _, section := range sections {
		_, err := tx.ExecContext(ctx, query, section, id)
		if err != nil {
			return emptyID, fmt.Errorf("%s: failed to insert section %s: %w", op, section, err)
		}
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return emptyID, fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetAllDepartments(ctx context.Context, institute string) ([]models.Department, error) {
	const op = "storage.postgresql.departments.GetAllDepartments"

	if err := s.SetSchema(ctx, institute); err != nil {
		return nil, err
	}

	var query string

	query = `SELECT id, name FROM departments`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var departments []models.Department
	for rows.Next() {
		var department models.Department

		err := rows.Scan(
			&department.ID,
			&department.Name,
		)

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

func (s *Storage) GetDepartmentID(ctx context.Context, institute string, name string) (int, error) {
	const op = "storage.postgresql.departments.GetDepartmnetID"

	if err := s.SetSchema(ctx, institute); err != nil {
		return emptyID, err
	}

	query := `SELECT id FROM departments WHERE name = $1`
	var id int
	err := s.db.QueryRowContext(ctx, query, name).Scan(&id)
	if err != nil {
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetSections(ctx context.Context, institute string, department string) ([]models.Section, error) {
	const op = "storage.postgresql.departments.GetSections"

	if err := s.SetSchema(ctx, institute); err != nil {
		return nil, err
	}

	parentID, err := s.GetDepartmentID(ctx, institute, department)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get department ID: %w", op, err)
	}

	query := `SELECT id, name FROM sections WHERE parent_id = $1`
	rows, err := s.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	var sections []models.Section
	for rows.Next() {
		var section models.Section
		err := rows.Scan(&section.ID, &section.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}
		sections = append(sections, section)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return sections, nil
}
func (s *Storage) DeleteDepartment(ctx context.Context, institute string, name string) error {
	const op = "storage.postgresql.departments.DeleteDepartment"

	if err := s.SetSchema(ctx, institute); err != nil {
		return err
	}

	query := `DELETE FROM departments WHERE name = $1`

	_, err := s.db.ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateDepartment(ctx context.Context, institute string, oldName string, name string, sections []string) error {
	const op = "storage.postgresql.departments.UpdateDepartment"

	// Начинаем транзакцию
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer tx.Rollback() // Убедимся, что транзакция будет откачена в случае ошибки

	if err := s.SetSchema(ctx, institute); err != nil {
		return err
	}

	// Удаляем старый отдел в рамках транзакции
	query := `DELETE FROM departments WHERE name = $1`
	_, err = tx.ExecContext(ctx, query, oldName)
	if err != nil {
		return fmt.Errorf("%s: failed to delete old department: %w", op, err)
	}

	// Создаем новый отдел в рамках транзакции
	var id int
	query = `INSERT INTO departments (name) VALUES ($1) RETURNING id`
	err = tx.QueryRowContext(ctx, query, name).Scan(&id)
	if err != nil {
		return fmt.Errorf("%s: failed to create new department: %w", op, err)
	}

	// Вставляем секции для нового отдела в рамках транзакции
	query = `INSERT INTO sections (name, parent_id) VALUES ($1, $2)`
	for _, section := range sections {
		_, err := tx.ExecContext(ctx, query, section, id)
		if err != nil {
			return fmt.Errorf("%s: failed to insert section %s: %w", op, section, err)
		}
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return nil
}
