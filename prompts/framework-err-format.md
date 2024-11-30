# RPC Framework Error Format Specification

## Wire Format

### Error Response Structure

[total length (4 bytes)][response ID (4 bytes)][error code (4 bytes)][error message (variable bytes)]

## Error Codes

const (
    ErrorCodeUnknown            uint32 = 0
    ErrorCodeMethodNotFound     uint32 = 1
    ErrorCodeInvalidRequest     uint32 = 2
    ErrorCodeMalformedRequest   uint32 = 3
    ErrorCodeInvalidMessageFormat uint32 = 4
    ErrorCodeInternalError      uint32 = 5
)

## Example

For an error response with:
- Request ID: 42
- Error Code: 1 (MethodNotFound)
- Message: "Calculator.Add not found"

The wire format would be:
```
[00 00 00 17]  // Length: 23 bytes (4 + 19)
[C0 00 00 2A]  // Response ID: 42 with MSB and second MSB set
[00 00 00 01]  // Error Code: MethodNotFound
[43 61 6C 63 75 6C 61 74 6F 72 2E 41 64 64 20 6E 6F 74 20 66 6F 75 6E 64]  // Message bytes
```

## Notes

1. This format is only for framework-level errors
2. Application-level errors should use protobuf-defined error responses
3. All multi-byte integers are in big-endian format
4. Message strings must be UTF-8 encoded

