package model

import (
	"net/mail"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Name == "" {
		errs["name"] = "is required"
	}
	if r.Email == "" {
		errs["email"] = "is required"
	} else if _, err := mail.ParseAddress(r.Email); err != nil {
		errs["email"] = "must be a valid email address"
	}
	if r.Password == "" {
		errs["password"] = "is required"
	} else if len(r.Password) < 8 {
		errs["password"] = "must be at least 8 characters"
	}

	return errs
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Email == "" {
		errs["email"] = "is required"
	} else if _, err := mail.ParseAddress(r.Email); err != nil {
		errs["email"] = "must be a valid email address"
	}
	if r.Password == "" {
		errs["password"] = "is required"
	}

	return errs
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
