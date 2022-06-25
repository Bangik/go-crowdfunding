package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	RegiserUser(input RegisterUserInput) (User, error)
	LoginUser(input LoginUserInput) (User, error)
	CheckEmailUnique(input CheckEmailInput) (bool, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) *service {
	return &service{repository}
}

// register user
func (s *service) RegiserUser(input RegisterUserInput) (User, error) {
	user := User{}
	user.Name = input.Name
	user.Occupation = input.Occupation
	user.Email = input.Email
	PasswordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.MinCost)
	if err != nil {
		return user, err
	}

	user.PasswordHash = string(PasswordHash)
	user.Role = "user"

	newUser, err := s.repository.Save(user)
	if err != nil {
		return newUser, err
	}

	return newUser, nil
}

// login user
func (s *service) LoginUser(input LoginUserInput) (User, error) {
	user, err := s.repository.FindByEmail(input.Email)
	if err != nil {
		return user, err
	}

	if user.ID == 0 {
		return user, errors.New("user not found with this email")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *service) CheckEmailUnique(input CheckEmailInput) (bool, error) {
	user, err := s.repository.FindByEmail(input.Email)
	if err != nil {
		return false, err
	}

	if user.ID == 0 {
		return true, nil
	}

	return false, nil
}
