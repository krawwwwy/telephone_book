package models

import "time"

// User представляет информацию о пользователе
type User struct {
	ID          int       `json:"id,omitempty"`
	Surname     string    `json:"surname"`
	Name        string    `json:"name"`
	MiddleName  string    `json:"middle_name,omitempty"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	Cabinet     string    `json:"cabinet,omitempty"`
	Position    string    `json:"position,omitempty"`
	Department  string    `json:"department"`
	Section     string    `json:"section,omitempty"`
	BirthDate   time.Time `json:"birth_date,omitempty"`
	Description string    `json:"description,omitempty"`
	Photo       []byte    `json:"photo,omitempty"`
}

// EmptyUser представляет пустого пользователя для возврата в случае ошибок
var EmptyUser = User{}
