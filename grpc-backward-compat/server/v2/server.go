package v2

import (
	"context"
	"fmt"
	"log"
	"sync"

	userv2 "grpc-backward-compat/proto/v2"
)

type Server struct {
	userv2.UnimplementedUserServiceServer
	users map[string]*userv2.User
	mu    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		users: make(map[string]*userv2.User),
	}
}

func (s *Server) GetUser(ctx context.Context, req *userv2.GetUserRequest) (*userv2.GetUserResponse, error) {
	log.Printf("[V2 Server] GetUser called for user_id: %s", req.UserId)

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[req.UserId]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", req.UserId)
	}

	log.Printf("[V2 Server] Returning user: id=%s, name=%s, email=%s, phone=%s", user.UserId, user.Name, user.Email, user.Phone)
	return &userv2.GetUserResponse{User: user}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *userv2.CreateUserRequest) (*userv2.CreateUserResponse, error) {
	log.Printf("[V2 Server] CreateUser called with name=%s, email=%s, phone=%s", req.Name, req.Email, req.Phone)

	userId := fmt.Sprintf("user_%d", len(s.users)+1)
	user := &userv2.User{
		UserId: userId,
		Name:   req.Name,
		Email:  req.Email,
		Phone:  req.Phone,
	}

	s.mu.Lock()
	s.users[userId] = user
	s.mu.Unlock()

	log.Printf("[V2 Server] Created user: id=%s, name=%s, email=%s, phone=%s", user.UserId, user.Name, user.Email, user.Phone)
	log.Printf("[V2 Server] Note: V2 server handles 'phone' field - if V1 client sends request, phone will be empty")

	return &userv2.CreateUserResponse{User: user}, nil
}
