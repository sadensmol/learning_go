package main

import (
	"log"
	"net"

	serverv2 "grpc-backward-compat/server/v2"
	userv2 "grpc-backward-compat/proto/v2"

	"google.golang.org/grpc"
)

const (
	port = ":50052"
)

func main() {
	log.Println("=== Scenario 2 Server: V2 Server (new proto WITH 'phone' field) ===")
	log.Println()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	userv2.RegisterUserServiceServer(s, serverv2.NewServer())

	log.Printf("[V2 Server] Listening on %s", port)
	log.Printf("[V2 Server] Ready to accept requests from V1 clients")
	log.Println()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
