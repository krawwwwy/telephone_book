package models

type Service struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}
