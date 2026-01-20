package v1

import (
	"context"
	"fmt"
	"log"
	"sync"

	userv1 "grpc-backward-compat/proto/v1"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	users map[string]*userv1.User
	mu    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		users: make(map[string]*userv1.User),
	}
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	log.Printf("[V1 Server] GetUser called for user_id: %s", req.UserId)

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[req.UserId]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", req.UserId)
	}

	log.Printf("[V1 Server] Returning user: id=%s, name=%s, email=%s", user.UserId, user.Name, user.Email)
	return &userv1.GetUserResponse{User: user}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	log.Printf("[V1 Server] CreateUser called with name=%s, email=%s", req.Name, req.Email)

	userId := fmt.Sprintf("user_%d", len(s.users)+1)
	user := &userv1.User{
		UserId: userId,
		Name:   req.Name,
		Email:  req.Email,
	}

	s.mu.Lock()
	s.users[userId] = user
	s.mu.Unlock()

	log.Printf("[V1 Server] Created user: id=%s, name=%s, email=%s", user.UserId, user.Name, user.Email)
	log.Printf("[V1 Server] Note: V1 server does not handle 'phone' field - it will be ignored if sent by V2 client")

	return &userv1.CreateUserResponse{User: user}, nil
}
