package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"time"

	"github.com/akhil4chelsia/grpc-go-microservice/calculator/calcpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
}

func (s *server) Sum(ctx context.Context, req *calcpb.CalcRequest) (*calcpb.CalcResponse, error) {
	fmt.Println("Calculating sum...")
	return &calcpb.CalcResponse{
		Result: req.X + req.Y,
	}, nil
}

func (s *server) PrimeNumberDecomposition(req *calcpb.PrimeNumberDecompositionRequest, stream calcpb.CalcService_PrimeNumberDecompositionServer) error {
	num := float64(req.Number)
	var k float64 = 2
	for num > 1 {
		if math.Mod(num, k) == 0 {
			stream.Send(&calcpb.PrimeNumberDecompositionResponse{
				Factor: float32(k),
			})
			num = num / k
		} else {
			k = k + 1
		}
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (s *server) CalcAverage(stream calcpb.CalcService_CalcAverageServer) error {
	fmt.Println("Calculating average with stream of numbers.")
	var sum float32 = 0
	var count float32 = 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&calcpb.CalcAverageResponse{
				Average: float32(sum / count),
			})
		}
		if err != nil {
			log.Fatal("Error while receiving stream.")
		}
		sum = sum + float32(req.GetNumber())
		count = count + 1
	}

}

func (s *server) FindMax(stream calcpb.CalcService_FindMaxServer) error {

	max := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("error while receiving stream %v", err)
		}

		if req.GetNumber() > int32(max) {
			max = int(req.GetNumber())
			sendErr := stream.Send(&calcpb.FindMaxResponse{
				Max: float64(max),
			})
			if sendErr != nil {
				log.Fatalf("Error while sending response %s", sendErr)
			}
		}
	}
}

func (s *server) SquareRoot(ctx context.Context, req *calcpb.SquareRootRequest) (*calcpb.SquareRootResponse, error) {

	num := float64(req.GetNumber())

	if num < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Negative number received %v", num),
		)
	}

	return &calcpb.SquareRootResponse{
		NumberRoot: math.Sqrt(num),
	}, nil
}

func main() {
	fmt.Println("Starting calc server...")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	checkError(err, "Error while starting listner.")

	s := grpc.NewServer()
	calcpb.RegisterCalcServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		checkError(err, "Failed to start grpc server.")
	}

}

func checkError(err error, message string) {
	if err != nil {
		log.Printf("%s , %v", message, err)
	}
}
