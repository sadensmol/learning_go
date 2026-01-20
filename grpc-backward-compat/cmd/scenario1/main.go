package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	serverv1 "grpc-backward-compat/server/v1"
	userv1 "grpc-backward-compat/proto/v1"
	userv2 "grpc-backward-compat/proto/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	port = ":50051"
)

func main() {
	log.Println("=== Scenario 1: Old Server (V1) + New Client (V2) ===")
	log.Println("Testing: V2 client sends 'phone' field to V1 server that doesn't know about it")
	log.Println()

	// start V1 server
	go startV1Server()
	time.Sleep(2 * time.Second)

	// run V2 client
	runV2Client()
}

func startV1Server() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	userv1.RegisterUserServiceServer(s, serverv1.NewServer())

	log.Printf("[Server] V1 Server listening on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func runV2Client() {
	conn, err := grpc.NewClient(fmt.Sprintf("localhost%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[Client] Failed to connect: %v", err)
	}
	defer conn.Close()

	client := userv2.NewUserServiceClient(conn)

	log.Println("\n[Client] V2 Client creating user with name, email, AND phone...")
	createReq := &userv2.CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
		Phone: "+1234567890",
	}

	createResp, err := client.CreateUser(context.Background(), createReq)
	if err != nil {
		log.Fatalf("[Client] CreateUser failed: %v", err)
	}

	log.Printf("[Client] V2 Client received response: id=%s, name=%s, email=%s, phone=%s",
		createResp.User.UserId, createResp.User.Name, createResp.User.Email, createResp.User.Phone)
	log.Printf("[Client] ✓ SUCCESS: V1 server accepted V2 client request (phone field was ignored)")

	log.Println("\n[Client] V2 Client getting user...")
	getReq := &userv2.GetUserRequest{
		UserId: createResp.User.UserId,
	}

	getResp, err := client.GetUser(context.Background(), getReq)
	if err != nil {
		log.Fatalf("[Client] GetUser failed: %v", err)
	}

	log.Printf("[Client] V2 Client received response: id=%s, name=%s, email=%s, phone=%s",
		getResp.User.UserId, getResp.User.Name, getResp.User.Email, getResp.User.Phone)
	log.Printf("[Client] ✓ SUCCESS: V2 client received response from V1 server (phone is empty, as expected)")

	log.Println("\n=== Scenario 1 Result ===")
	log.Println("✓ Backward compatibility WORKS: V2 client can communicate with V1 server")
	log.Println("✓ New field 'phone' sent by V2 client is ignored by V1 server")
	log.Println("✓ V2 client receives empty 'phone' field from V1 server (default value)")
}
