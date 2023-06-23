package services

import (
	"errors"
	"time"

	"github.com/seanomeara96/gates/repositories"
)

type ProductService struct {
	userRepository *repositories.ProductRepository
}

func NewUserService(userRepository *repositories.ProductRepository) *ProductService {
	return &UserService{userRepository: userRepository}
}

func (s *ProductService) CreateUser(name string, email string) (*models.User, error) {
	// Validate input parameters
	if name == "" || email == "" {
		return nil, errors.New("name and email are required")
	}

	// Check if the email is already registered
	existingUser, err := s.productRepository.GetByEmail(email)
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
	err = s.productRepository.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *ProductService) GetUserByID(userID int) (*models.User, error) {
	// Fetch the user from the database
	user, err := s.productRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Other methods for user-related operations (e.g., UpdateUser, DeleteUser, etc.)
