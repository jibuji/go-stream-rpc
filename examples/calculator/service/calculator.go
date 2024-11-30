package calculator

import (
	"context"
	"fmt"

	proto "github.com/jibuji/go-stream-rpc/examples/calculator/proto"
)

type CalculatorService struct {
	proto.UnimplementedCalculatorServer
}

func (s *CalculatorService) Add(ctx context.Context, req *proto.AddRequest) *proto.AddResponse {
	result := req.A + req.B
	fmt.Printf("Server handling Add: %d + %d = %d\n", req.A, req.B, result)
	return &proto.AddResponse{Result: result}
}

func (s *CalculatorService) Multiply(ctx context.Context, req *proto.MultiplyRequest) *proto.MultiplyResponse {
	result := req.A * req.B
	fmt.Printf("Server handling Multiply: %d * %d = %d\n", req.A, req.B, result)
	return &proto.MultiplyResponse{Result: result}
}
