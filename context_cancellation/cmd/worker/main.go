package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// workerServer implements the WorkerService
type workerServer struct {
	UnimplementedWorkerServiceServer
	db *sql.DB
}

// DoWork implements the DoWork RPC method
func (s *workerServer) DoWork(ctx context.Context, req *WorkRequest) (*WorkResponse, error) {
	taskID := req.TaskId
	log.Printf("[WORKER] Received work request: task_id=%s, data=%s", taskID, req.Data)

	// Start database transaction on worker server
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[WORKER] Failed to start transaction: %v", err)
		return nil, status.Error(codes.Internal, "failed to start transaction")
	}

	// Ensure transaction is rolled back if we don't commit
	defer func() {
		if tx != nil {
			log.Printf("[WORKER] Rolling back transaction for task_id=%s", taskID)
			tx.Rollback()
		}
	}()

	// Insert a record into worker database
	_, err = tx.ExecContext(ctx, "INSERT INTO worker_tasks (task_id, data) VALUES (?, ?)", taskID, req.Data)
	if err != nil {
		log.Printf("[WORKER] Failed to insert task: %v", err)
		return nil, status.Error(codes.Internal, "failed to insert task")
	}

	log.Printf("[WORKER] Inserted task record for task_id=%s", taskID)

	// Simulate long-running work with periodic context checking
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	endTime := time.Now().Add(3 * time.Second)

	for {
		select {
		case <-ctx.Done():
			// Context was cancelled - rollback transaction
			log.Printf("[WORKER] Context cancelled for task_id=%s: %v", taskID, ctx.Err())
			if err := tx.Rollback(); err != nil {
				// If transaction was already rolled back by the driver due to context cancellation, that's ok
				if err == sql.ErrTxDone {
					log.Printf("[WORKER] ❌ TRANSACTION ROLLED BACK for task_id=%s (automatically by driver)", taskID)
				} else {
					log.Printf("[WORKER] ⚠️  Failed to rollback transaction for task_id=%s: %v", taskID, err)
				}
			} else {
				log.Printf("[WORKER] ❌ TRANSACTION ROLLED BACK for task_id=%s", taskID)
			}
			tx = nil // Prevent double rollback in defer
			return nil, status.Error(codes.Canceled, "work cancelled")

		case <-ticker.C:
			if time.Now().After(endTime) {
				// Work completed successfully - commit transaction
				if err := tx.Commit(); err != nil {
					log.Printf("[WORKER] Failed to commit transaction: %v", err)
					return nil, status.Error(codes.Internal, "failed to commit")
				}
				tx = nil // Prevent rollback in defer

				log.Printf("[WORKER] ✅ TRANSACTION COMMITTED for task_id=%s", taskID)
				return &WorkResponse{
					Success: true,
					Message: fmt.Sprintf("Work completed for task %s", taskID),
				}, nil
			}
		}
	}
}

func main() {
	port := flag.Int("port", 50051, "gRPC server port")
	dbPath := flag.String("db", "./test_worker.db", "Database path")
	flag.Parse()

	log.Printf("[WORKER] Starting gRPC Worker server on port %d with database %s", *port, *dbPath)

	// Initialize worker database
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("[WORKER] Failed to open database: %v", err)
	}
	defer db.Close()

	// Create worker_tasks table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS worker_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("[WORKER] Failed to create table: %v", err)
	}

	log.Printf("[WORKER] Database initialized successfully")

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("[WORKER] Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	RegisterWorkerServiceServer(grpcServer, &workerServer{db: db})

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Printf("[WORKER] Shutting down gracefully...")
		grpcServer.GracefulStop()
	}()

	log.Printf("[WORKER] gRPC Worker server listening on :%d", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[WORKER] Server error: %v", err)
	}
}
