package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/akhil4chelsia/grpc-go-microservice/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {
	tls := false
	var opts = grpc.WithInsecure()
	if tls {
		certFile := "ssl/ca.crt"
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Failed to load ssl certificate %v", sslErr)
			return
		}
		opts = grpc.WithTransportCredentials(creds)
	}
	cc, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("Could not connect to server. %v", err)
	}
	defer cc.Close()
	c := greetpb.NewGreetServiceClient(cc)
	doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	//doBidirectionalStreaming(c)

	//doGreetWithDeadline(c, 5*time.Second) // should complete
	//doGreetWithDeadline(c, 1*time.Second) // should fail
}

func doUnary(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Akhil",
			LastName:  "Muralidharan",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling greet rpc, %v", err)
	}
	fmt.Printf("Response : %v\n", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Doing server streaming rpc")
	req := &greetpb.GreetManyTimesRequest{
		FirstName: "Amal",
	}
	stream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling greetmanytimes . %v", err)
	}
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while reading stream. %v", err)
		}
		fmt.Printf("Result: %s\n", msg.GetResult())
	}

}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Doing client streaming rpc")

	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			FirstName: "Akhil",
		},
		&greetpb.LongGreetRequest{
			FirstName: "Amal",
		},
		&greetpb.LongGreetRequest{
			FirstName: "Namitha",
		},
		&greetpb.LongGreetRequest{
			FirstName: "Ameya",
		},
		&greetpb.LongGreetRequest{
			FirstName: "Unni",
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatal("Error while calling longgreet")
	}

	for _, req := range requests {
		fmt.Printf("Sending request for %s\n", req.GetFirstName())
		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("Error while receiving response")
	}

	fmt.Printf("Response: %s\n", res.GetResult())
}

func doBidirectionalStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Doing bi directional streaming rpc")

	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			FirstName: "Akhil",
		},
		&greetpb.GreetEveryoneRequest{
			FirstName: "Amal",
		},
		&greetpb.GreetEveryoneRequest{
			FirstName: "Namitha",
		},
		&greetpb.GreetEveryoneRequest{
			FirstName: "Ameya",
		},
		&greetpb.GreetEveryoneRequest{
			FirstName: "Unni",
		},
	}

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while opening stream %v", err)
	}
	waitc := make(chan struct{})
	go func() {
		for _, req := range requests {
			fmt.Printf("Sending request for %s\n", req.GetFirstName())
			err := stream.Send(req)
			if err != nil {
				log.Fatalf("Error while sending request %v", err)
			}
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving stream %v", err)
				break
			}
			fmt.Printf("Response: %s\n", res.GetResult())
		}
		close(waitc)
	}()

	<-waitc

}

func doGreetWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Printf("Calling doGreetWithDeadline with timeout : %v \n", timeout)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancel()

	res, err := c.GreetWithDeadline(ctx, &greetpb.GreetWithDeadlineRequest{
		FirstName: "Akhil",
	})

	if err != nil {

		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout hit , deadline exceeded.")
			} else {
				fmt.Println("unexpected error.")
			}
		} else {
			log.Fatalf("Error while calling rpc function -GreetWithDeadline %v", statusErr)
		}

	}

	fmt.Printf("Result: %v\n", res.GetResult())
}
