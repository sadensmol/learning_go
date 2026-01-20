# Distributed Transaction Context Cancellation Demo

**Two completely separate OS processes** communicating via HTTP and gRPC, proving that Go's context can manage distributed transaction cancellation across process boundaries - **exactly like Kubernetes/Docker microservices!**

## Architecture

```
┌─────────────────────────────────┐       ┌──────────────────────────────────┐
│   HTTP Echo Server (Process 1) │       │ gRPC Worker Server (Process 2)   │
│   cmd/echo/main.go              │       │ cmd/worker/main.go               │
├─────────────────────────────────┤       ├──────────────────────────────────┤
│ Port: 8080                      │       │ Port: 50051                      │
│ Database: test_echo.db          │       │ Database: test_worker.db         │
│ Table: echo_requests            │       │ Table: worker_tasks              │
│                                 │       │                                  │
│ 1. Receives HTTP request        │       │ 4. Receives gRPC call            │
│ 2. Starts DB transaction        │       │ 5. Starts DB transaction         │
│ 3. Calls gRPC Worker ──────────────────>│ 6. Simulates 3 sec work          │
│ 7. Commits/rollbacks            │       │ 8. Commits/rollbacks             │
└─────────────────────────────────┘       └──────────────────────────────────┘
        SEPARATE OS PROCESS                      SEPARATE OS PROCESS
```

## Key Features

1. **SEPARATE OS PROCESSES**: Echo and Worker run as independent processes (different PIDs)
2. **Real Databases**: Each process has its own SQLite database with transactions
3. **Context Propagation**: HTTP request context flows across process boundary to gRPC
4. **Automatic Rollback**: When context cancelled, **both** processes rollback their transactions
5. **Proof via Database**: Tests verify by checking actual database records after processes communicate
6. **Emulates K8s/Docker**: Exactly how microservices work in production!

## Project Structure

```
context_cancellation/
├── cmd/
│   ├── echo/
│   │   ├── main.go          # HTTP Echo server executable
│   │   ├── worker.pb.go     # Protobuf code
│   │   └── worker_grpc.pb.go
│   └── worker/
│       ├── main.go          # gRPC Worker server executable
│       ├── worker.pb.go     # Protobuf code
│       └── worker_grpc.pb.go
├── worker.proto             # gRPC service definition
├── worker.pb.go             # Generated protobuf (root)
├── worker_grpc.pb.go        # Generated protobuf (root)
└── main_test.go             # Tests that launch both processes
```

## Running the Application

### Start gRPC Worker Server (Process 1)
```bash
cd cmd/worker
go run main.go worker.pb.go worker_grpc.pb.go -port=50051 -db=../../test_worker.db
```

### Start HTTP Echo Server (Process 2)
```bash
cd cmd/echo
go run main.go worker.pb.go worker_grpc.pb.go -http-port=8080 -grpc-port=50051 -db=../../test_echo.db
```

### Test It
```bash
# This will complete successfully (both transactions commit)
curl 'http://localhost:8080/echo?request_id=test-001&message=hello'

# Check the databases
sqlite3 test_echo.db "SELECT * FROM echo_requests;"
sqlite3 test_worker.db "SELECT * FROM worker_tasks;"
```

## Running Tests

The test suite launches **BOTH servers as separate OS processes** using `exec.Command()`:

```bash
go test -v
```

### Test Output - Showing Separate Processes

```
=== RUN   TestDistributedTransactionCancellation
Starting SEPARATE OS PROCESSES: Echo on :52613, Worker on :52612
Worker server process started with PID: 9870  ← SEPARATE PROCESS!
Echo server process started with PID: 9908    ← SEPARATE PROCESS!

[WORKER] Starting gRPC Worker server on port 52612
[ECHO] Starting HTTP Echo server on port 52613

=== Test Case: Context Cancellation (Separate OS Processes) ===
1. Starting HTTP request...
   [ECHO] Inserted echo record
   [WORKER] Inserted task record
2. Cancelling HTTP request context...
   [ECHO] ❌ TRANSACTION ROLLED BACK
   [WORKER] ❌ TRANSACTION ROLLED BACK
4. Verifying database state...
   Echo database records: 0  ✅
   Worker database records: 0  ✅

=== Test Case: Successful Completion (Separate OS Processes) ===
5. Starting HTTP request that will complete...
   [WORKER] ✅ TRANSACTION COMMITTED
   [ECHO] ✅ TRANSACTION COMMITTED
7. Verifying database state...
   Echo database records: 1  ✅
   Worker database records: 1  ✅

=== Test Summary ===
✓ HTTP Echo and gRPC Worker running as SEPARATE OS PROCESSES
✓ Context cancellation rolls back BOTH transactions
✓ Successful completion commits BOTH transactions
✓ Emulates real Kubernetes/Docker microservices architecture!
--- PASS: TestDistributedTransactionCancellation

=== RUN   TestHTTPConnectionClosePropagation
Starting SEPARATE OS PROCESSES: Echo on :52627, Worker on :52626
Worker server process started with PID: 9974  ← DIFFERENT PID!
Echo server process started with PID: 10012   ← DIFFERENT PID!

=== Testing TCP Connection Close → gRPC Cancellation (Separate Processes) ===
1. Starting HTTP request...
2. Forcefully closing TCP connection...
   [ECHO] ❌ TRANSACTION ROLLED BACK
   [WORKER] ❌ TRANSACTION ROLLED BACK
4. Worker database records: 0  ✅

=== Summary ===
✓ TCP connection closure propagated across SEPARATE OS PROCESSES
✓ gRPC worker in different process detected cancellation
✓ Transaction rolled back successfully
--- PASS: TestHTTPConnectionClosePropagation
```

## How It Works

### Process Communication Flow
```
HTTP Client
    ↓ HTTP Request (with context)
Process 1: Echo Server (PID: XXXX)
    ↓ Opens transaction in echo.db
    ↓ Inserts record
    ↓ gRPC Call with same context →→→→→→ Process 2: Worker Server (PID: YYYY)
    ↓                                        ↓ Opens transaction in worker.db
    ↓                                        ↓ Inserts record
    ↓                                        ↓ Simulates 3 sec work
    ↓                                        ↓ Monitors ctx.Done()
    ↓←←←←←←← Response/Error ←←←←←←←←←←←←←←←←↓
    ↓ Commits/Rollbacks transaction          ↓ Commits/Rollbacks transaction
```

### Cancelled Request Flow (Separate Processes)
1. HTTP request arrives at Echo server (Process 1, PID: 1234)
2. Echo opens DB transaction, inserts record
3. Echo calls gRPC Worker (Process 2, PID: 5678) **via network**
4. Worker opens DB transaction, inserts record
5. **HTTP context is cancelled** (client disconnects)
6. Context cancellation propagates **across process boundary** via gRPC
7. Worker detects `ctx.Done()`, rolls back transaction ❌
8. Echo receives cancellation error, rolls back transaction ❌
9. **Both databases are empty** (verified by test)

## The Answer: YES!

**Can Go's context manage distributed transactions across separate OS processes?**

**ABSOLUTELY YES!** This application proves it:

1. ✅ Two completely separate OS processes (different PIDs)
2. ✅ Separate databases (test_echo.db and test_worker.db)
3. ✅ gRPC communication across process boundary
4. ✅ Context propagates from HTTP → gRPC → Worker process
5. ✅ When context cancelled, **both processes** rollback their transactions
6. ✅ Verified by actual database state (not test mocks)
7. ✅ Emulates real Kubernetes/Docker microservices!

## Why This Matters

In production microservices (Kubernetes/Docker):
- Each service runs as a **separate container/process**
- Services communicate via HTTP/gRPC **over network**
- Each service has its **own database**
- Context cancellation must propagate across **process boundaries**

**This demo proves it works perfectly with Go's context mechanism!**

## Database Schema

### Echo Server (`test_echo.db`)
```sql
CREATE TABLE echo_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Worker Server (`test_worker.db`)
```sql
CREATE TABLE worker_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Learning Points

- **Context crosses process boundaries**: Go's context mechanism works perfectly across gRPC
- **BeginTx with context**: Using `db.BeginTx(ctx, nil)` ties transactions to context lifecycle
- **Defer for safety**: `defer tx.Rollback()` ensures rollback on any error or cancellation
- **Periodic checking**: Long-running work must check `ctx.Done()` regularly
- **Real proof**: Separate OS processes + database verification = bulletproof
- **Production ready**: This pattern works for real distributed systems
- **No 2PC needed**: For cancellation-only scenarios, context is sufficient

## Limitations

This demonstrates **transaction cancellation only** - not:
- Distributed commits (2-phase commit)
- Cross-database ACID guarantees
- Automatic rollback on one server if the other fails after commit

For full distributed transactions, use:
- 2-Phase Commit (2PC)
- Saga pattern
- Distributed transaction coordinators (XA transactions)

But for **cancellation propagation across microservices**, Go's context is perfect!
