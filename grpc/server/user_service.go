package server

import (
	"context"
	"go-gin-project/models"
	pb "go-gin-project/proto"
	"go-gin-project/services"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGrpcService struct {
	pb.UnimplementedUserServiceServer
	userService *services.UserService
}

func NewUserGrpcService(userService *services.UserService) *UserGrpcService {
	return &UserGrpcService{
		userService: userService,
	}
}

func (s *UserGrpcService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	user := &models.User{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	createdUser, err := s.userService.Create(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &pb.UserResponse{
		Id:        strconv.FormatUint(uint64(createdUser.ID), 10),
		Email:     createdUser.Email,
		Name:      createdUser.Name,
		CreatedAt: createdUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: createdUser.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserGrpcService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := s.userService.Get(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.UserResponse{
		Id:        strconv.FormatUint(uint64(user.ID), 10),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserGrpcService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	updates := &models.User{}
	if req.Email != nil {
		updates.Email = *req.Email
	}
	if req.Name != nil {
		updates.Name = *req.Name
	}

	updatedUser, err := s.userService.Update(req.Id, updates)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &pb.UserResponse{
		Id:        strconv.FormatUint(uint64(updatedUser.ID), 10),
		Email:     updatedUser.Email,
		Name:      updatedUser.Name,
		CreatedAt: updatedUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: updatedUser.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserGrpcService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	err := s.userService.Delete(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}
