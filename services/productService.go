package services

import (
	"errors"
	"time"

	"github.com/your_project/models"
	"github.com/your_project/repositories"
)

type UserService struct {
	userRepository *repositories.UserRepository
}

func NewUserService(userRepository *repositories.UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) CreateUser(name string, email string) (*models.User, error) {
	// Validate input parameters
	if name == "" || email == "" {
		return nil, errors.New("name and email are required")
	}

	// Check if the email is already registered
	existingUser, err := s.userRepository.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email is already registered")
	}

	// Create a new user instance
	user := &models.User{
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}

	// Save the user to the database
	err = s.userRepository.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByID(userID int) (*models.User, error) {
	// Fetch the user from the database
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Other methods for user-related operations (e.g., UpdateUser, DeleteUser, etc.)
