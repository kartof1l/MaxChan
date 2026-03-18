package models

import "time"

type User struct {
	ID        int
	Email     string
	Password  string
	IsAdmin   bool
	CreatedAt time.Time
}

type Product struct {
	ID          int
	Name        string
	Price       int
	Description string
	FullDesc    string
	ImageURL    string
	CreatedAt   time.Time
}

type Review struct {
	ID        int
	ProductID int
	UserID    int
	UserName  string
	Rating    int
	Comment   string
	CreatedAt time.Time
}

type Session struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
}
