package calculator

import (
	"context"
	"fmt"
	proto "stream-rpc/examples/calculator/proto"
)

type CalculatorService struct {
}

func (s *CalculatorService) Add(ctx context.Context, req *proto.AddRequest) (*proto.AddResponse, error) {
	result := req.A + req.B
	fmt.Printf("Server handling Add: %d + %d = %d\n", req.A, req.B, result)
	return &proto.AddResponse{Result: result}, nil
}

func (s *CalculatorService) Multiply(ctx context.Context, req *proto.MultiplyRequest) (*proto.MultiplyResponse, error) {
	result := req.A * req.B
	fmt.Printf("Server handling Multiply: %d * %d = %d\n", req.A, req.B, result)
	return &proto.MultiplyResponse{Result: result}, nil
}
