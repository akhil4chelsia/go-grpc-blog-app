package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/akhil4chelsia/grpc-go-microservice/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
}

func (s *server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Println("Invoked Green() on server.")
	firstname := req.GetGreeting().GetFirstName()
	result := "Hello, " + firstname
	res := &greetpb.GreetResponse{
		Result: result,
	}
	return res, nil
}

func (s *server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Print("Greet many times invoked.")
	first_name := req.FirstName
	for i := 0; i < 10; i++ {
		res := &greetpb.GreetManyTimesResponse{
			Result: "hello, " + first_name + strconv.Itoa(i) + " times.",
		}
		stream.Send(res)
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (s *server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Println("Invoking LongGreet on server.")
	result := ""
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading stream.")
		}
		result = result + "Hello " + req.GetFirstName() + ", "
	}
}

func (s *server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatal("error while receiving stream")
		}
		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: "Hello, " + req.GetFirstName(),
		})

		if sendErr != nil {
			log.Fatalf("Error while sending response, %v", sendErr)
			return sendErr
		}
	}
}

func (s *server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("client cancelled the request.")
			return nil, status.Error(codes.Canceled, "Deadline exceeded for client request.")
		}
		time.Sleep(1 * time.Second)
	}
	return &greetpb.GreetWithDeadlineResponse{
		Result: "Hello " + req.GetFirstName(),
	}, nil
}

func main() {

	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to start listner. %v", err)
	}
	tls := false
	opts := []grpc.ServerOption{}
	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Failed loading ssl certificates %v", sslErr)
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)
	greetpb.RegisterGreetServiceServer(s, &server{})
	reflection.Register(s)

	fmt.Println("Starting geeting server...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to start server. %v", err)
	}
}
