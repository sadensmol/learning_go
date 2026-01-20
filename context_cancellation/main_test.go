package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Helper functions for database operations
func countEchoRecords(t *testing.T, dbPath string) int {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open echo database: %v", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM echo_requests").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count echo records: %v", err)
	}
	return count
}

func countWorkerRecords(t *testing.T, dbPath string) int {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open worker database: %v", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM worker_tasks").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count worker records: %v", err)
	}
	return count
}

func clearDatabases(t *testing.T, echoDbPath, workerDbPath string) {
	// Clear echo database
	echoDb, err := sql.Open("sqlite3", echoDbPath)
	if err != nil {
		t.Fatalf("Failed to open echo database: %v", err)
	}
	defer echoDb.Close()
	_, err = echoDb.Exec("DELETE FROM echo_requests")
	if err != nil {
		t.Fatalf("Failed to clear echo_requests: %v", err)
	}

	// Clear worker database
	workerDb, err := sql.Open("sqlite3", workerDbPath)
	if err != nil {
		t.Fatalf("Failed to open worker database: %v", err)
	}
	defer workerDb.Close()
	_, err = workerDb.Exec("DELETE FROM worker_tasks")
	if err != nil {
		t.Fatalf("Failed to clear worker_tasks: %v", err)
	}
}

// ProcessWithLogs wraps exec.Cmd with log capturing
type ProcessWithLogs struct {
	*exec.Cmd
	logs *bytes.Buffer
}

// startWorkerProcess starts the gRPC worker server as a separate OS process
func startWorkerProcess(t *testing.T, port int, dbPath string) *ProcessWithLogs {
	// Build the worker server first
	t.Log("Building worker server binary...")
	buildCmd := exec.Command("go", "build", "-o", "worker_server_bin", "./cmd/worker")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build worker server: %v\nOutput: %s", err, output)
	}

	// Start the worker server process
	t.Logf("Starting worker server as OS process on port %d...", port)
	cmd := exec.Command("./worker_server_bin", fmt.Sprintf("-port=%d", port), fmt.Sprintf("-db=%s", dbPath))

	// Capture logs to buffer while also showing them
	logBuffer := &bytes.Buffer{}
	multiWriter := io.MultiWriter(os.Stdout, logBuffer)
	cmd.Stdout = multiWriter
	cmd.Stderr = multiWriter

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start worker process: %v", err)
	}

	// Wait for worker to be ready
	time.Sleep(500 * time.Millisecond)

	t.Logf("Worker server process started with PID: %d", cmd.Process.Pid)
	return &ProcessWithLogs{Cmd: cmd, logs: logBuffer}
}

// startEchoProcess starts the HTTP echo server as a separate OS process
func startEchoProcess(t *testing.T, httpPort, grpcPort int, dbPath string) *exec.Cmd {
	// Build the echo server first
	t.Log("Building echo server binary...")
	buildCmd := exec.Command("go", "build", "-o", "echo_server_bin", "./cmd/echo")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build echo server: %v\nOutput: %s", err, output)
	}

	// Start the echo server process
	t.Logf("Starting echo server as OS process on port %d...", httpPort)
	cmd := exec.Command("./echo_server_bin",
		fmt.Sprintf("-http-port=%d", httpPort),
		fmt.Sprintf("-grpc-port=%d", grpcPort),
		fmt.Sprintf("-db=%s", dbPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start echo process: %v", err)
	}

	// Wait for echo server to be ready
	time.Sleep(500 * time.Millisecond)

	t.Logf("Echo server process started with PID: %d", cmd.Process.Pid)
	return cmd
}

// TestDistributedTransactionCancellation tests distributed transactions with separate OS processes
func TestDistributedTransactionCancellation(t *testing.T) {
	grpcPort := findAvailablePort(t)
	httpPort := findAvailablePort(t)
	echoDbPath := "./test_echo_dist.db"
	workerDbPath := "./test_worker_dist.db"

	// Clean up old databases
	os.Remove(echoDbPath)
	os.Remove(workerDbPath)
	defer os.Remove(echoDbPath)
	defer os.Remove(workerDbPath)

	t.Logf("Starting SEPARATE OS PROCESSES: Echo on :%d, Worker on :%d", httpPort, grpcPort)

	// Start gRPC Worker server as separate OS process
	workerProc := startWorkerProcess(t, grpcPort, workerDbPath)
	defer func() {
		t.Log("Killing worker process...")
		workerProc.Process.Kill()
		workerProc.Wait()
		os.Remove("./worker_server_bin")
	}()

	// Start HTTP Echo server as separate OS process
	echoCmd := startEchoProcess(t, httpPort, grpcPort, echoDbPath)
	defer func() {
		t.Log("Killing echo process...")
		echoCmd.Process.Kill()
		echoCmd.Wait()
		os.Remove("./echo_server_bin")
	}()

	t.Log("\n=== Test Case: Context Cancellation (Separate OS Processes) ===")

	// Create HTTP client with cancellable context
	client := &http.Client{Transport: &http.Transport{}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create request
	requestID := "cancelled-request-001"
	url := fmt.Sprintf("http://localhost:%d/echo?request_id=%s&message=test", httpPort, requestID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Channel to signal when the request completes
	requestDone := make(chan error, 1)

	// Start the request in a goroutine
	t.Log("1. Starting HTTP request...")
	go func() {
		resp, err := client.Do(req)
		if err != nil {
			requestDone <- err
			return
		}
		defer resp.Body.Close()
		_, _ = io.ReadAll(resp.Body)
		requestDone <- nil
	}()

	// Wait to ensure both transactions have started
	time.Sleep(500 * time.Millisecond)

	t.Log("2. Cancelling HTTP request context...")
	cancel()

	// Wait for request to complete
	select {
	case err := <-requestDone:
		if err != nil {
			t.Logf("3. HTTP request failed as expected: %v", err)
		} else {
			t.Fatal("HTTP request should have failed")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("HTTP request timeout")
	}

	// Give time for transactions to rollback
	time.Sleep(1 * time.Second)

	// Check database state - both should be empty (transactions rolled back)
	t.Log("4. Verifying database state after cancellation...")
	echoCount := countEchoRecords(t, echoDbPath)
	workerCount := countWorkerRecords(t, workerDbPath)

	t.Logf("   Echo database records: %d", echoCount)
	t.Logf("   Worker database records: %d", workerCount)

	if echoCount != 0 {
		t.Errorf("❌ Echo transaction was NOT rolled back! Found %d records (expected 0)", echoCount)
	} else {
		t.Log("   ✅ Echo transaction was rolled back correctly")
	}

	if workerCount != 0 {
		t.Errorf("❌ Worker transaction was NOT rolled back! Found %d records (expected 0)", workerCount)
	} else {
		t.Log("   ✅ Worker transaction was rolled back correctly")
	}

	t.Log("\n=== Test Case: Successful Completion (Separate OS Processes) ===")

	// Clear databases again
	clearDatabases(t, echoDbPath, workerDbPath)

	// Now test successful case
	requestID = "successful-request-001"
	url = fmt.Sprintf("http://localhost:%d/echo?request_id=%s&message=success", httpPort, requestID)

	t.Log("5. Starting HTTP request that will complete successfully...")
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make successful request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	t.Logf("6. Response: %s", string(body))

	// Give time for transactions to commit
	time.Sleep(1 * time.Second)

	// Check database state - both should have records
	t.Log("7. Verifying database state after successful completion...")
	echoCount = countEchoRecords(t, echoDbPath)
	workerCount = countWorkerRecords(t, workerDbPath)

	t.Logf("   Echo database records: %d", echoCount)
	t.Logf("   Worker database records: %d", workerCount)

	if echoCount != 1 {
		t.Errorf("❌ Echo transaction should be committed! Found %d records (expected 1)", echoCount)
	} else {
		t.Log("   ✅ Echo transaction was committed correctly")
	}

	if workerCount != 1 {
		t.Errorf("❌ Worker transaction should be committed! Found %d records (expected 1)", workerCount)
	} else {
		t.Log("   ✅ Worker transaction was committed correctly")
	}

	t.Log("\n=== Test Summary ===")
	t.Log("✓ HTTP Echo and gRPC Worker running as SEPARATE OS PROCESSES")
	t.Log("✓ Context cancellation rolls back BOTH transactions")
	t.Log("✓ Successful completion commits BOTH transactions")
	t.Log("✓ Emulates real Kubernetes/Docker microservices architecture!")
}

// TestHTTPConnectionClosePropagation tests TCP connection closure with separate OS processes
func TestHTTPConnectionClosePropagation(t *testing.T) {
	grpcPort := findAvailablePort(t)
	httpPort := findAvailablePort(t)
	echoDbPath := "./test_echo_conn.db"
	workerDbPath := "./test_worker_conn.db"

	// Clean up old databases
	os.Remove(echoDbPath)
	os.Remove(workerDbPath)
	defer os.Remove(echoDbPath)
	defer os.Remove(workerDbPath)

	t.Logf("Starting SEPARATE OS PROCESSES: Echo on :%d, Worker on :%d", httpPort, grpcPort)

	// Start gRPC Worker server as separate OS process WITH LOG CAPTURING
	workerProc := startWorkerProcess(t, grpcPort, workerDbPath)
	defer func() {
		workerProc.Process.Kill()
		workerProc.Wait()
		os.Remove("./worker_server_bin")
	}()

	// Start HTTP Echo server as separate OS process
	echoCmd := startEchoProcess(t, httpPort, grpcPort, echoDbPath)
	defer func() {
		echoCmd.Process.Kill()
		echoCmd.Wait()
		os.Remove("./echo_server_bin")
	}()

	t.Log("\n=== Testing TCP Connection Close → gRPC Cancellation (Separate Processes) ===")

	// Create a custom dialer that tracks connections
	var activeConn net.Conn
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			activeConn = conn
			return conn, nil
		},
	}
	client := &http.Client{Transport: transport}

	requestID := "conn-close-test"
	url := fmt.Sprintf("http://localhost:%d/echo?request_id=%s&message=test", httpPort, requestID)
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	requestDone := make(chan error, 1)

	t.Log("1. Starting HTTP request...")
	go func() {
		resp, err := client.Do(req)
		if err != nil {
			requestDone <- err
			return
		}
		defer resp.Body.Close()
		_, _ = io.ReadAll(resp.Body)
		requestDone <- nil
	}()

	time.Sleep(500 * time.Millisecond)

	t.Log("2. Forcefully closing TCP connection...")
	if activeConn != nil {
		activeConn.Close()
	}

	select {
	case err := <-requestDone:
		if err != nil {
			t.Logf("3. HTTP request failed as expected: %v", err)
		} else {
			t.Fatal("HTTP request should have failed")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("HTTP request timeout")
	}

	time.Sleep(1 * time.Second)

	// Check database state
	workerCount := countWorkerRecords(t, workerDbPath)
	t.Logf("4. Worker database records after connection close: %d", workerCount)

	if workerCount != 0 {
		t.Errorf("❌ gRPC call was NOT cancelled! Found %d worker records (expected 0)", workerCount)
	} else {
		t.Log("   ✅ gRPC call was cancelled (transaction rolled back)")
	}

	// PROOF: Check worker process logs for cancellation message
	t.Log("\n5. Verifying gRPC cancellation in worker process logs...")
	logs := workerProc.logs.String()

	if strings.Contains(logs, "Context cancelled for task_id="+requestID) {
		t.Log("   ✅ Found context cancellation log in worker process!")
		t.Logf("   Log snippet: 'Context cancelled for task_id=%s'", requestID)
	} else {
		t.Error("   ❌ Did NOT find context cancellation in worker logs!")
	}

	if strings.Contains(logs, "TRANSACTION ROLLED BACK for task_id="+requestID) {
		t.Log("   ✅ Found transaction rollback log in worker process!")
		t.Logf("   Log snippet: 'TRANSACTION ROLLED BACK for task_id=%s'", requestID)
	} else {
		t.Error("   ❌ Did NOT find transaction rollback in worker logs!")
	}

	t.Log("\n=== Summary ===")
	t.Log("✓ TCP connection closure propagated across SEPARATE OS PROCESSES")
	t.Log("✓ gRPC worker in different process detected cancellation (VERIFIED IN LOGS)")
	t.Log("✓ Transaction rolled back successfully (VERIFIED IN LOGS + DATABASE)")
}

func findAvailablePort(t *testing.T) int {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}
