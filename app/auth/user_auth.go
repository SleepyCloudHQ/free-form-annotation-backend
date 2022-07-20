package auth

import (
	"backend/app/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserAuth struct {
	DB *gorm.DB
}

func NewUserAuth(db *gorm.DB) *UserAuth {
	return &UserAuth{DB: db}
}

func (a *UserAuth) CreateUser(email string, password string, role models.UserRole) (*models.User, error) {
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), 8)
	if hashErr != nil {
		return nil, hashErr
	}

	user := &models.User{Email: email, Password: string(hashedPassword), Role: role}
	result := a.DB.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func (a *UserAuth) CheckUserPassword(email string, password string) (*models.User, error) {
	user := &models.User{}
	result := a.DB.First(user, "email = ?", email)
	if result.Error != nil {
		return nil, result.Error
	}

	compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if compareErr != nil {
		return nil, compareErr
	}

	return user, nil
}
