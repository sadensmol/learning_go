package main

import (
	"log"
	"net"

	serverv1 "grpc-backward-compat/server/v1"
	userv1 "grpc-backward-compat/proto/v1"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func main() {
	log.Println("=== Scenario 1 Server: V1 Server (old proto without 'phone' field) ===")
	log.Println()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	userv1.RegisterUserServiceServer(s, serverv1.NewServer())

	log.Printf("[V1 Server] Listening on %s", port)
	log.Printf("[V1 Server] Ready to accept requests from V2 clients")
	log.Println()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
