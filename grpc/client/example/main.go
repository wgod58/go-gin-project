package main

import (
	"log"

	"go-gin-project/grpc/client"
)

func main() {
	userClient, err := client.NewUserClient("localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer userClient.Close()

	// Example: Create a new user
	user, err := userClient.CreateUser("test@example.com", "password123", "Test User")
	if err != nil {
		log.Fatalf("Could not create user: %v", err)
	}
	log.Printf("Created user: %v", user)

	// Example: Get the user
	fetchedUser, err := userClient.GetUser(user.Id)
	if err != nil {
		log.Fatalf("Could not get user: %v", err)
	}
	log.Printf("Fetched user: %v", fetchedUser)
}
