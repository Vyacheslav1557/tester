package models

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Role int32

type User struct {
	Id             int32     `db:"id"`
	Username       string    `db:"username"`
	HashedPassword string    `db:"hashed_pwd"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
	Role           Role      `db:"role"`
}

type UserCreation struct {
	Username string
	Password string
	Role     Role
}

func (u *UserCreation) HashPassword() error {
	hpwd, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hpwd)
	return nil
}

func (user *User) IsSamePwd(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return false
	}
	return true
}

type UsersListFilters struct {
	PageSize int32
	Page     int32
	Role     *Role
	Username *string
	Order    *int32
}

func (f UsersListFilters) Offset() int32 {
	return (f.Page - 1) * f.PageSize
}

type UsersList struct {
	Users      []*User
	Pagination Pagination
}

type UserUpdate struct {
	Username *string
	Role     *Role
}

const (
	RoleGuest   Role = -1
	RoleStudent Role = 0
	RoleTeacher Role = 1
	RoleAdmin   Role = 2
)

type Grant struct {
	Action   string
	Resource string
}
