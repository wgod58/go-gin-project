package client

import (
	"context"
	"fmt"
	"time"

	pb "go-gin-project/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserClient(address string) (*UserClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &UserClient{
		client: pb.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *UserClient) Close() error {
	return c.conn.Close()
}

func (c *UserClient) CreateUser(email, password, name string) (*pb.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return c.client.CreateUser(ctx, &pb.CreateUserRequest{
		Email:    email,
		Password: password,
		Name:     name,
	})
}

func (c *UserClient) GetUser(id string) (*pb.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return c.client.GetUser(ctx, &pb.GetUserRequest{Id: id})
}

func (c *UserClient) UpdateUser(id, email, name string) (*pb.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return c.client.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:    id,
		Email: &email,
		Name:  &name,
	})
}

func (c *UserClient) DeleteUser(id string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}
