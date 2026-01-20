# gRPC Backward Compatibility Test Project

This project demonstrates backward and forward compatibility in gRPC by testing communication between clients and servers using different proto file versions.

## Project Structure

```
.
├── proto/
│   ├── v1/          # Old proto (without 'phone' field)
│   └── v2/          # New proto (with 'phone' field)
├── server/
│   ├── v1/          # Server implementation using V1 proto
│   └── v2/          # Server implementation using V2 proto
└── cmd/
    ├── scenario1-server/  # V1 server for scenario 1
    ├── scenario1-client/  # V2 client for scenario 1
    ├── scenario2-server/  # V2 server for scenario 2
    └── scenario2-client/  # V1 client for scenario 2
```

## Proto File Versions

### V1 Proto (Old - proto/v1/user.proto)
```protobuf
message User {
  string user_id = 1;
  string name = 2;
  string email = 3;
}
```

### V2 Proto (New - proto/v2/user.proto)
```protobuf
message User {
  string user_id = 1;
  string name = 2;
  string email = 3;
  string phone = 4;  // Additional field
}
```

## Test Scenarios

### Scenario 1: Old Server (V1) + New Client (V2)

**Setup:**
- Server uses V1 proto (without 'phone' field)
- Client uses V2 proto (with 'phone' field)

**Test:**
- V2 client sends CreateUser request with phone="+1234567890"
- V1 server processes request (ignores unknown 'phone' field)
- V1 server responds without 'phone' field
- V2 client receives response with empty 'phone' field

**Result:** ✓ PASSED - Backward compatibility works

**Logs:**

Client output:
```
2026/01/20 12:20:52 === Scenario 1 Client: V2 Client (new proto WITH 'phone' field) ===
2026/01/20 12:20:53 [V2 Client] Creating user with name, email, AND phone...
2026/01/20 12:20:53 [V2 Client] Sending request with phone='+1234567890' (field not in V1 server)
2026/01/20 12:20:53 [V2 Client] ✓ Response received: id=user_1, name=John Doe, email=john@example.com, phone=''
2026/01/20 12:20:53 [V2 Client] ✓ Phone field is EMPTY (as expected - V1 server doesn't know about it)
2026/01/20 12:20:53 ✓ V2 client successfully communicated with V1 server
2026/01/20 12:20:53 ✓ 'phone' field sent by V2 client was ignored by V1 server (backward compatible)
```

Server output:
```
2026/01/20 12:20:50 === Scenario 1 Server: V1 Server (old proto without 'phone' field) ===
2026/01/20 12:20:50 [V1 Server] Listening on :50051
2026/01/20 12:20:53 [V1 Server] CreateUser called with name=John Doe, email=john@example.com
2026/01/20 12:20:53 [V1 Server] Created user: id=user_1, name=John Doe, email=john@example.com
2026/01/20 12:20:53 [V1 Server] Note: V1 server does not handle 'phone' field - it will be ignored if sent by V2 client
```

### Scenario 2: New Server (V2) + Old Client (V1)

**Setup:**
- Server uses V2 proto (with 'phone' field)
- Client uses V1 proto (without 'phone' field)

**Test:**
- V1 client sends CreateUser request (no 'phone' field)
- V2 server processes request (phone defaults to empty string)
- V2 server responds with empty 'phone' field
- V1 client receives response (ignores unknown 'phone' field)

**Result:** ✓ PASSED - Forward compatibility works

**Logs:**

Client output:
```
2026/01/20 12:21:38 === Scenario 2 Client: V1 Client (old proto WITHOUT 'phone' field) ===
2026/01/20 12:21:39 [V1 Client] Creating user with name and email (NO phone field)
2026/01/20 12:21:39 [V1 Client] V1 client doesn't even know about 'phone' field
2026/01/20 12:21:39 [V1 Client] ✓ Response received: id=user_1, name=Jane Smith, email=jane@example.com
2026/01/20 12:21:39 [V1 Client] ✓ Request accepted by V2 server (forward compatible)
2026/01/20 12:21:39 ✓ V1 client successfully communicated with V2 server
2026/01/20 12:21:39 ✓ Missing 'phone' field from V1 client was accepted by V2 server (forward compatible)
```

Server output:
```
2026/01/20 12:21:36 === Scenario 2 Server: V2 Server (new proto WITH 'phone' field) ===
2026/01/20 12:21:36 [V2 Server] Listening on :50052
2026/01/20 12:21:39 [V2 Server] CreateUser called with name=Jane Smith, email=jane@example.com, phone=
2026/01/20 12:21:39 [V2 Server] Created user: id=user_1, name=Jane Smith, email=jane@example.com, phone=
2026/01/20 12:21:39 [V2 Server] Note: V2 server handles 'phone' field - if V1 client sends request, phone will be empty
```

## How to Run

### Prerequisites
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Generate Proto Files
```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/v1/user.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/v2/user.proto
```

### Build Binaries
```bash
go build -o cmd/scenario1-server/server cmd/scenario1-server/main.go
go build -o cmd/scenario1-client/client cmd/scenario1-client/main.go
go build -o cmd/scenario2-server/server cmd/scenario2-server/main.go
go build -o cmd/scenario2-client/client cmd/scenario2-client/main.go
```

### Run Scenario 1 (V1 Server + V2 Client)
```bash
# Terminal 1 - Start V1 server
./cmd/scenario1-server/server

# Terminal 2 - Run V2 client
./cmd/scenario1-client/client
```

### Run Scenario 2 (V2 Server + V1 Client)
```bash
# Terminal 1 - Start V2 server
./cmd/scenario2-server/server

# Terminal 2 - Run V1 client
./cmd/scenario2-client/client
```

## Key Findings

### ✓ Backward Compatibility Works
- New clients can communicate with old servers
- Unknown fields sent by new clients are ignored by old servers
- Old servers respond without new fields
- New clients handle missing fields (use default values)

### ✓ Forward Compatibility Works
- Old clients can communicate with new servers
- Old clients don't send new fields
- New servers accept requests with missing fields (use defaults)
- New servers can send responses with new fields
- Old clients ignore unknown fields in responses

### Important Notes

1. **Same Package Name Required**: Both proto versions must use the same package name for the service to be compatible. In this project, both use `package user;`

2. **Field Numbers**: Field numbers must remain stable. Never reuse field numbers for different purposes.

3. **Default Values**: In proto3, missing fields default to their zero values (empty string for strings, 0 for numbers, etc.)

4. **Separate Binaries**: Client and server must be separate binaries to avoid proto message name conflicts when both versions are imported.

## Conclusion

This project successfully demonstrates that gRPC provides excellent backward and forward compatibility when adding new fields to proto messages. As long as:
- The service and package names remain the same
- Field numbers are not reused
- New fields are optional (proto3 default)

...then old and new versions can interoperate seamlessly.
