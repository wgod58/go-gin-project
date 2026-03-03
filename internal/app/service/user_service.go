package service

import (
	"fmt"
	"time"

	"go-gin-project/internal/pkg/model"
)

type UserService struct {
	repo  model.UserRepository
	cache model.CacheService
}

func NewUserService(repo model.UserRepository, cache model.CacheService) *UserService {
	return &UserService{repo: repo, cache: cache}
}

func (s *UserService) Create(user *model.User) (*model.User, error) {
	return s.repo.Create(user)
}

func (s *UserService) Get(id string) (*model.User, error) {
	cacheKey := fmt.Sprintf("user:%s", id)

	var user model.User
	if err := s.cache.Get(cacheKey, &user); err == nil {
		return &user, nil
	}

	found, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	s.cache.Set(cacheKey, found, 5*time.Minute) //nolint:errcheck
	return found, nil
}

func (s *UserService) Update(id string, data *model.User) (*model.User, error) {
	updated, err := s.repo.Update(id, data)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	s.cache.Delete(fmt.Sprintf("user:%s", id)) //nolint:errcheck
	return updated, nil
}

func (s *UserService) Delete(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	s.cache.Delete(fmt.Sprintf("user:%s", id)) //nolint:errcheck
	return nil
}
