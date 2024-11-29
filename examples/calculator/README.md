# Calculator Example

This example demonstrates how to use the RPC framework to create a simple calculator service.

## Structure

- `proto/` - Protocol Buffer definitions
- `service/` - Calculator service implementation
- `cmd/` - Server and client implementations

## Running the Example

1. Start the server: 

```bash
go run cmd/server/main.go
```

2. In another terminal, run the client:

```bash
go run cmd/client/main.go
```
    