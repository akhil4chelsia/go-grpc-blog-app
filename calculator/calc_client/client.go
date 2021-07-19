package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/akhil4chelsia/grpc-go-microservice/calculator/calcpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	checkError(err, "Error while connecting to server.")
	c := calcpb.NewCalcServiceClient(cc)
	req := &calcpb.CalcRequest{
		X: 10,
		Y: 20,
	}
	res, err := c.Sum(context.Background(), req)
	checkError(err, "Error while invoking Sum method")
	fmt.Printf("Result: %v\n", res.Result)
	//doServerStreaming(c)
	//calAverageWithClientStreaming(c)
	//getMaxNumber(c)
	calcSquareRootError(c)
}

func doServerStreaming(c calcpb.CalcServiceClient) {
	req2 := &calcpb.PrimeNumberDecompositionRequest{
		Number: 120,
	}

	stream, err := c.PrimeNumberDecomposition(context.Background(), req2)
	checkError(err, "Error while calling remote method-PrimeNumberDecomposition ")

	for {
		v, err := stream.Recv()
		if err == io.EOF {
			break
		}
		checkError(err, "Error while recieving stream.")
		fmt.Printf("Factor : %v\n", v.GetFactor())
	}
}

func calAverageWithClientStreaming(c calcpb.CalcServiceClient) {

	stream, err := c.CalcAverage(context.Background())
	checkError(err, "Error while callign calc average")
	numbers := []int64{
		1, 2, 3, 4,
	}

	for _, n := range numbers {
		fmt.Printf("Sending number %v\n", n)
		stream.Send(&calcpb.CalcAverageRequest{
			Number: n,
		})
	}
	res, err := stream.CloseAndRecv()
	checkError(err, "Error while receiving response")
	fmt.Printf("Average: %v\n", res.GetAverage())
}

func getMaxNumber(c calcpb.CalcServiceClient) {

	stream, err := c.FindMax(context.Background())
	checkError(err, "Error while creating stream")

	numbers := []int32{1, 5, 3, 6, 2, 20}
	waitc := make(chan struct{})
	go func() {

		for _, n := range numbers {
			fmt.Printf("Sending number %v \n", n)
			sendErr := stream.Send(&calcpb.FindmaxRequest{
				Number: n,
			})
			if sendErr != nil {
				log.Fatalf("Error while sending req %v", sendErr)
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
				log.Fatalf("error while recieving stream %v", err)
			}

			fmt.Printf("Max Number : %v\n", res.GetMax())
		}
		close(waitc)
	}()

	<-waitc

}

func calcSquareRootError(c calcpb.CalcServiceClient) {
	res, err := c.SquareRoot(context.Background(), &calcpb.SquareRootRequest{
		Number: -2,
	})

	if err != nil {
		resErr, ok := status.FromError(err)
		if ok {
			//GRPC custom error
			fmt.Printf("Code : %v, Message: %v\n", resErr.Code(), resErr.Message())
			if resErr.Code() == codes.InvalidArgument {
				fmt.Println("We sent a negative number")
			}

		} else {
			log.Fatalf("Unexpected error %v", resErr)
		}
	}
	fmt.Printf("Result : %v \n", res.GetNumberRoot())
}

func checkError(err error, message string) {
	if err != nil {
		log.Printf("%s , %v", message, err)
	}
}
