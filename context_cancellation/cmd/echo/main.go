package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Global database for echo server
var echoDb *sql.DB

// initEchoDatabase initializes the echo server SQLite database
func initEchoDatabase(dbPath string) error {
	var err error

	// Initialize echo server database
	echoDb, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open echo database: %v", err)
	}

	// Create echo_requests table
	_, err = echoDb.Exec(`
		CREATE TABLE IF NOT EXISTS echo_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			request_id TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create echo_requests table: %v", err)
	}

	log.Println("[ECHO] Database initialized successfully")
	return nil
}

// httpEchoHandler handles HTTP requests and makes gRPC calls
func httpEchoHandler(grpcClient WorkerServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.URL.Query().Get("request_id")
		if requestID == "" {
			requestID = fmt.Sprintf("req-%d", time.Now().Unix())
		}

		message := r.URL.Query().Get("message")
		if message == "" {
			message = "Hello"
		}

		log.Printf("[ECHO] Received HTTP request: request_id=%s, message=%s", requestID, message)

		// Start database transaction on echo server
		ctx := r.Context()
		tx, err := echoDb.BeginTx(ctx, nil)
		if err != nil {
			log.Printf("[ECHO] Failed to start transaction: %v", err)
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}

		// Ensure transaction is rolled back if we don't commit
		defer func() {
			if tx != nil {
				log.Printf("[ECHO] Rolling back transaction for request_id=%s", requestID)
				tx.Rollback()
			}
		}()

		// Insert a record into echo database
		_, err = tx.ExecContext(ctx, "INSERT INTO echo_requests (request_id, message) VALUES (?, ?)", requestID, message)
		if err != nil {
			log.Printf("[ECHO] Failed to insert request: %v", err)
			http.Error(w, "Failed to insert request", http.StatusInternalServerError)
			return
		}

		log.Printf("[ECHO] Inserted echo record for request_id=%s", requestID)

		// Pass request_id to gRPC via metadata
		ctx = metadata.AppendToOutgoingContext(ctx, "request-id", requestID)

		// Call gRPC worker service
		log.Printf("[ECHO] Calling gRPC worker service for request_id=%s", requestID)
		resp, err := grpcClient.DoWork(ctx, &WorkRequest{
			TaskId: requestID,
			Data:   message,
		})

		if err != nil {
			log.Printf("[ECHO] gRPC call failed for request_id=%s: %v", requestID, err)
			if rbErr := tx.Rollback(); rbErr != nil {
				// If transaction was already rolled back by the driver due to context cancellation, that's ok
				if rbErr == sql.ErrTxDone {
					log.Printf("[ECHO] ❌ TRANSACTION ROLLED BACK for request_id=%s (automatically by driver)", requestID)
				} else {
					log.Printf("[ECHO] ⚠️  Failed to rollback transaction for request_id=%s: %v", requestID, rbErr)
				}
			} else {
				log.Printf("[ECHO] ❌ TRANSACTION ROLLED BACK for request_id=%s", requestID)
			}
			tx = nil // Prevent double rollback in defer
			http.Error(w, fmt.Sprintf("Worker failed: %v", err), http.StatusInternalServerError)
			return
		}

		// gRPC call succeeded - commit echo transaction
		if err := tx.Commit(); err != nil {
			log.Printf("[ECHO] Failed to commit transaction: %v", err)
			http.Error(w, "Failed to commit", http.StatusInternalServerError)
			return
		}
		tx = nil // Prevent rollback in defer

		log.Printf("[ECHO] ✅ TRANSACTION COMMITTED for request_id=%s", requestID)
		log.Printf("[ECHO] Response from worker: %s", resp.Message)
		fmt.Fprintf(w, "Success: %s\n", resp.Message)
	}
}

func main() {
	httpPort := flag.Int("http-port", 8080, "HTTP server port")
	grpcPort := flag.Int("grpc-port", 50051, "gRPC server port (to connect to)")
	dbPath := flag.String("db", "./test_echo.db", "Database path")
	flag.Parse()

	log.Printf("[ECHO] Starting HTTP Echo server on port %d", *httpPort)
	log.Printf("[ECHO] Will connect to gRPC Worker at localhost:%d", *grpcPort)

	// Initialize echo database
	if err := initEchoDatabase(*dbPath); err != nil {
		log.Fatalf("[ECHO] Failed to initialize database: %v", err)
	}
	defer echoDb.Close()

	// Create gRPC client connection to worker service
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", *grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("[ECHO] Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	grpcClient := NewWorkerServiceClient(conn)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", httpEchoHandler(grpcClient))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *httpPort),
		Handler: mux,
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Printf("[ECHO] Shutting down gracefully...")
		server.Shutdown(context.Background())
	}()

	log.Printf("[ECHO] HTTP Echo server listening on :%d", *httpPort)
	log.Printf("[ECHO] Try: curl 'http://localhost:%d/echo?request_id=test-001&message=hello'", *httpPort)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[ECHO] Server error: %v", err)
	}
}
