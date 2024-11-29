package calculator

import (
	"context"
	pb "stream-rpc/examples/calculator/proto"
)

type CalculatorService struct {
}

func (s *CalculatorService) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	return &pb.AddResponse{Result: req.A + req.B}, nil
}

func (s *CalculatorService) Multiply(ctx context.Context, req *pb.MultiplyRequest) (*pb.MultiplyResponse, error) {
	return &pb.MultiplyResponse{Result: req.A * req.B}, nil
}
