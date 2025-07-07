package models

type User struct {
	Surname     string `json:"surname"`
	Name        string `json:"name"`
	MiddleName  string `json:"middle_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Cabinet     string `json:"cabinet"`
	Position    string `json:"position"`
	Department  string `json:"department"`
}

var EmptyUser = User{}
