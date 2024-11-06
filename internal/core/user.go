package core

import (
	"context"
	"time"
)

type (
	User struct {
		ID                  int
		Username            string
		Email               *string
		FirstName           string
		LastName            string
		MiddleName          *string
		Pseudonym           string
		Telephone           *string
		PasswordHash        string
		IsEmailVerified     bool
		IsTelephoneVerified bool
		IsDeleted           bool
		CreatedAt           time.Time
		UpdatedAt           time.Time
	}

	UpdateUser struct {
		ID                  int
		Username            *string
		Email               *string
		FirstName           *string
		LastName            *string
		MiddleName          *string
		Pseudonym           *string
		Telephone           *string
		Password            *UpdatePassword
		IsEmailVerified     *bool
		IsTelephoneVerified *bool
	}

	UpdatePassword struct {
		OldPassword string
		NewPassword string
	}

	UserService interface {
		UpdateUser(ctx context.Context, user UpdateUser) (*User, error)
		DeleteUser(ctx context.Context, userID int) error
		GetUser(ctx context.Context, user User) (*User, error)
	}

	UserStore interface {
		AddUser(ctx context.Context, user User) (userID int, err error)
		GetUserByUsername(ctx context.Context, username string) (user *User, err error)
		GetUserByID(ctx context.Context, userID int) (user *User, err error)
		GetUserByEmail(ctx context.Context, email string) (user *User, err error)
		GetUserByTelephone(ctx context.Context, telephone string) (user *User, err error)
		UpdateUser(cxt context.Context, user UpdateUser) (*User, error)
		DeleteUser(ctx context.Context, userID int) error
	}
)
