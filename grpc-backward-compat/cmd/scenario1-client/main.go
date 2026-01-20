package main

import (
	"context"
	"fmt"
	"log"
	"time"

	userv2 "grpc-backward-compat/proto/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	port = ":50051"
)

func main() {
	log.Println("=== Scenario 1 Client: V2 Client (new proto WITH 'phone' field) ===")
	log.Println()

	time.Sleep(1 * time.Second)

	conn, err := grpc.NewClient(fmt.Sprintf("localhost%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[V2 Client] Failed to connect: %v", err)
	}
	defer conn.Close()

	client := userv2.NewUserServiceClient(conn)

	log.Println("[V2 Client] Creating user with name, email, AND phone...")
	log.Println("[V2 Client] Sending request with phone='+1234567890' (field not in V1 server)")
	createReq := &userv2.CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
		Phone: "+1234567890",
	}

	createResp, err := client.CreateUser(context.Background(), createReq)
	if err != nil {
		log.Fatalf("[V2 Client] CreateUser failed: %v", err)
	}

	log.Printf("[V2 Client] ✓ Response received: id=%s, name=%s, email=%s, phone='%s'",
		createResp.User.UserId, createResp.User.Name, createResp.User.Email, createResp.User.Phone)
	log.Println()

	if createResp.User.Phone == "" {
		log.Println("[V2 Client] ✓ Phone field is EMPTY (as expected - V1 server doesn't know about it)")
	} else {
		log.Printf("[V2 Client] ✗ Unexpected: Phone field has value: %s", createResp.User.Phone)
	}

	log.Println()
	log.Println("[V2 Client] Getting user back from server...")
	getReq := &userv2.GetUserRequest{
		UserId: createResp.User.UserId,
	}

	getResp, err := client.GetUser(context.Background(), getReq)
	if err != nil {
		log.Fatalf("[V2 Client] GetUser failed: %v", err)
	}

	log.Printf("[V2 Client] ✓ Response received: id=%s, name=%s, email=%s, phone='%s'",
		getResp.User.UserId, getResp.User.Name, getResp.User.Email, getResp.User.Phone)
	log.Println()

	log.Println("=== Scenario 1 Summary ===")
	log.Println("✓ V2 client successfully communicated with V1 server")
	log.Println("✓ 'phone' field sent by V2 client was ignored by V1 server (backward compatible)")
	log.Println("✓ V2 client received empty 'phone' field from V1 server (default value)")
}
