package parser

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"telephone-book/internal/domain/models"
	"time"

	"github.com/xuri/excelize/v2"
)

const layout = "2006-01-02"

func Excel(file multipart.File) ([]models.User, error) {
	// Читаем в память
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from sheet %s: %w", sheet, err)
	}

	var users []models.User

	// предполагаем, что первая строка — заголовки
	for i, row := range rows {
		if i == 0 {
			continue
		}

		birthDate, err := time.Parse(layout, getValue(row, 9))
		if err != nil {
			birthDate = time.Time{} // если не удалось, оставляем пустую дату
		}

		user := models.User{
			Surname:     getValue(row, 0),
			Name:        getValue(row, 1),
			MiddleName:  getValue(row, 2),
			Email:       getValue(row, 3),
			PhoneNumber: getValue(row, 4),
			Cabinet:     getValue(row, 5),
			Position:    getValue(row, 6),
			Department:  getValue(row, 7),
			Section:     getValue(row, 8),
			BirthDate:   birthDate,
			Description: getValue(row, 10),
		}

		users = append(users, user)
	}

	return users, nil
}

func getValue(row []string, index int) string {
	if len(row) > index {
		return row[index]
	}
	return ""
}
