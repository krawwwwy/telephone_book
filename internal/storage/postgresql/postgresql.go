package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func (s *Storage) SetSchema(ctx context.Context, institute string) error {
	var schema string
	switch institute {
	case "grafit", "графит", "Графит", "Grafit":
		schema = "grafit"
	case "giredmet", "Giredmet", "гиредмет", "Гиредмет":
		schema = "giredmet"
	default:
		return storage.ErrSchemaNotExist
	}
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`SET search_path TO %s`, pq.QuoteIdentifier(schema)))
	return err
}

func (s *Storage) Search(ctx context.Context, institute string, department string, section string, info string) ([]models.User, error) {
	const op = "storage.postgresql.Search"

	if err := s.SetSchema(ctx, institute); err != nil {
		return nil, err
	}

	// Разбиваем поисковую строку на слова
	words := strings.Fields(strings.TrimSpace(info))
	if len(words) == 0 {
		return []models.User{}, nil
	}

	var query string
	var args []interface{}

	// Строим условия поиска для каждого слова
	var searchConditions []string
	argIndex := 1

	for _, word := range words {
		// Каждое слово ищем по всем полям - только с начала слова
		wordCondition := fmt.Sprintf("(surname ILIKE $%d OR name ILIKE $%d OR middle_name ILIKE $%d OR email ILIKE $%d OR cabinet ILIKE $%d)",
			argIndex, argIndex, argIndex, argIndex, argIndex)
		searchConditions = append(searchConditions, wordCondition)
		args = append(args, word+"%") // Убираю % в начале - теперь ищем только с начала
		argIndex++
	}

	// Объединяем все условия через AND (все слова должны найтись)
	allSearchConditions := "(" + strings.Join(searchConditions, " AND ") + ")"

	if department == "" {
		// Если отдел не указан, возвращаем всех пользователей
		query = fmt.Sprintf(`SELECT id, surname, name, middle_name, email, phone_number, cabinet, position, department
			FROM workers
			WHERE %s
			ORDER BY surname, name`, allSearchConditions)
	} else {
		if section == "" {
			// Если отдел указан, фильтруем по нему
			query = fmt.Sprintf(`SELECT id, surname, name, middle_name, email, phone_number, cabinet, position, department
				FROM workers
				WHERE department = $%d AND %s
				ORDER BY surname, name`, argIndex, allSearchConditions)
			args = append(args, department)
		} else {
			// Если отдел и секция указаны, фильтруем по ним
			query = fmt.Sprintf(`SELECT id, surname, name, middle_name, email, phone_number, cabinet, position, department
				FROM workers
				WHERE department = $%d AND section = $%d AND %s
				ORDER BY surname, name`, argIndex, argIndex+1, allSearchConditions)
			args = append(args, department, section)
		}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var middleName, cabinet, position, department sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Surname,
			&user.Name,
			&middleName,
			&user.Email,
			&user.PhoneNumber,
			&cabinet,
			&position,
			&department,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		// Конвертируем NullString в обычные строки
		user.MiddleName = middleName.String
		user.Cabinet = cabinet.String
		user.Position = position.String
		user.Department = department.String

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return users, nil
}

func (s *Storage) CreateUserTx(
	ctx context.Context,
	tx *sql.Tx,
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

	err := tx.QueryRowContext(
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

func (s *Storage) ImportUsers(ctx context.Context, institute string, users []models.User) error {
	const op = "storage.postgresql.ImportUsers"

	if err := s.SetSchema(ctx, institute); err != nil {
		return err
	}

	if len(users) == 0 {
		return fmt.Errorf("%s: no users to import", op)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer tx.Rollback()

	for _, user := range users {
		_, err := s.CreateUserTx(
			ctx,
			tx,
			institute,
			user.Surname,
			user.Name,
			user.MiddleName,
			user.Email,
			user.PhoneNumber,
			user.Cabinet,
			user.Position,
			user.Department,
			user.Section,
			user.BirthDate,
			user.Description,
			nil,
		)
		if err != nil {
			return fmt.Errorf("%s: failed to create user %s: %w", op, user.Email, err)
		}
	}

	return tx.Commit()
}

func (s *Storage) Emergency(ctx context.Context) ([]models.Service, error) {
	const op = "storage.postgesql.Emergency"

	rows, err := s.db.QueryContext(ctx, `SELECT name, phone_number, email FROM public.main`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var service models.Service

		err := rows.Scan(
			&service.Name,
			&service.PhoneNumber,
			&service.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return services, nil
}

func (s *Storage) GetTodaysBirthdays(ctx context.Context, institute string) ([]models.User, error) {
	return s.getBirthdaysByOffset(ctx, institute, 0)
}

func (s *Storage) GetTomorrowsBirthdays(ctx context.Context, institute string) ([]models.User, error) {
	return s.getBirthdaysByOffset(ctx, institute, 1)
}

// Общая функция для смещения
func (s *Storage) getBirthdaysByOffset(ctx context.Context, institute string, dayOffset int) ([]models.User, error) {
	const op = "storage.postgresql.getBirthdaysByOffset"

	var schema string
	switch institute {
	case "grafit", "графит", "Графит", "Grafit":
		schema = "grafit"
	case "giredmet", "Giredmet", "гиредмет", "Гиредмет":
		schema = "giredmet"
	default:
		return nil, storage.ErrSchemaNotExist
	}

	// SQL: выбираем по смещению от текущей даты
	query := fmt.Sprintf(`
        SELECT id, surname, name, middle_name, email, phone_number, cabinet, position, department, birth_date
        FROM %s.workers
        WHERE birth_date IS NOT NULL 
          AND EXTRACT(MONTH FROM birth_date) = EXTRACT(MONTH FROM CURRENT_DATE + INTERVAL '%d day')
          AND EXTRACT(DAY FROM birth_date) = EXTRACT(DAY FROM CURRENT_DATE + INTERVAL '%d day')
        ORDER BY surname, name
    `, schema, dayOffset, dayOffset)

	rows, err := s.db.QueryContext(ctx, query)
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
			&user.BirthDate,
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
