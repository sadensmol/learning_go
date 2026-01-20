package main

import (
	"context"
	"fmt"
	"log"
	"time"

	userv1 "grpc-backward-compat/proto/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	port = ":50052"
)

func main() {
	log.Println("=== Scenario 2 Client: V1 Client (old proto WITHOUT 'phone' field) ===")
	log.Println()

	time.Sleep(1 * time.Second)

	conn, err := grpc.NewClient(fmt.Sprintf("localhost%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[V1 Client] Failed to connect: %v", err)
	}
	defer conn.Close()

	client := userv1.NewUserServiceClient(conn)

	log.Println("[V1 Client] Creating user with name and email (NO phone field)")
	log.Println("[V1 Client] V1 client doesn't even know about 'phone' field")
	createReq := &userv1.CreateUserRequest{
		Name:  "Jane Smith",
		Email: "jane@example.com",
	}

	createResp, err := client.CreateUser(context.Background(), createReq)
	if err != nil {
		log.Fatalf("[V1 Client] CreateUser failed: %v", err)
	}

	log.Printf("[V1 Client] ✓ Response received: id=%s, name=%s, email=%s",
		createResp.User.UserId, createResp.User.Name, createResp.User.Email)
	log.Println()
	log.Println("[V1 Client] ✓ Request accepted by V2 server (forward compatible)")

	log.Println()
	log.Println("[V1 Client] Getting user back from server...")
	getReq := &userv1.GetUserRequest{
		UserId: createResp.User.UserId,
	}

	getResp, err := client.GetUser(context.Background(), getReq)
	if err != nil {
		log.Fatalf("[V1 Client] GetUser failed: %v", err)
	}

	log.Printf("[V1 Client] ✓ Response received: id=%s, name=%s, email=%s",
		getResp.User.UserId, getResp.User.Name, getResp.User.Email)
	log.Println()

	log.Println("=== Scenario 2 Summary ===")
	log.Println("✓ V1 client successfully communicated with V2 server")
	log.Println("✓ Missing 'phone' field from V1 client was accepted by V2 server (forward compatible)")
	log.Println("✓ V1 client ignores 'phone' field sent by V2 server (doesn't know about it)")
}
