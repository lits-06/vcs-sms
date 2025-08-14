package repository

import "github.com/lits-06/vcs-sms/domain/entity"

type UserRepository interface {
	CreateUser(user *entity.User) error
	GetUserByID(id string) (*entity.User, error)
	GetUserByEmail(email string) (*entity.User, error)
	UpdateUser(user *entity.User) error
	DeleteUser(id string) error
}
