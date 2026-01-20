package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	serverv2 "grpc-backward-compat/server/v2"
	userv1 "grpc-backward-compat/proto/v1"
	userv2 "grpc-backward-compat/proto/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	port = ":50052"
)

func main() {
	log.Println("=== Scenario 2: New Server (V2) + Old Client (V1) ===")
	log.Println("Testing: V1 client doesn't send 'phone' field to V2 server that expects it")
	log.Println()

	// start V2 server
	go startV2Server()
	time.Sleep(2 * time.Second)

	// run V1 client
	runV1Client()
}

func startV2Server() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	userv2.RegisterUserServiceServer(s, serverv2.NewServer())

	log.Printf("[Server] V2 Server listening on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func runV1Client() {
	conn, err := grpc.NewClient(fmt.Sprintf("localhost%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[Client] Failed to connect: %v", err)
	}
	defer conn.Close()

	client := userv1.NewUserServiceClient(conn)

	log.Println("\n[Client] V1 Client creating user with name and email (NO phone)...")
	createReq := &userv1.CreateUserRequest{
		Name:  "Jane Smith",
		Email: "jane@example.com",
	}

	createResp, err := client.CreateUser(context.Background(), createReq)
	if err != nil {
		log.Fatalf("[Client] CreateUser failed: %v", err)
	}

	log.Printf("[Client] V1 Client received response: id=%s, name=%s, email=%s",
		createResp.User.UserId, createResp.User.Name, createResp.User.Email)
	log.Printf("[Client] ✓ SUCCESS: V2 server accepted V1 client request (phone field was empty)")

	log.Println("\n[Client] V1 Client getting user...")
	getReq := &userv1.GetUserRequest{
		UserId: createResp.User.UserId,
	}

	getResp, err := client.GetUser(context.Background(), getReq)
	if err != nil {
		log.Fatalf("[Client] GetUser failed: %v", err)
	}

	log.Printf("[Client] V1 Client received response: id=%s, name=%s, email=%s",
		getResp.User.UserId, getResp.User.Name, getResp.User.Email)
	log.Printf("[Client] ✓ SUCCESS: V1 client received response from V2 server (phone field is ignored)")

	log.Println("\n=== Scenario 2 Result ===")
	log.Println("✓ Forward compatibility WORKS: V1 client can communicate with V2 server")
	log.Println("✓ Missing 'phone' field from V1 client is accepted by V2 server (defaults to empty)")
	log.Println("✓ V1 client ignores additional 'phone' field sent by V2 server")
}
