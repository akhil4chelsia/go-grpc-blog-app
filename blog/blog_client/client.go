package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/akhil4chelsia/grpc-go-microservice/blog/blogpb"
	"google.golang.org/grpc"
)

func main() {

	var opts = grpc.WithInsecure()

	cc, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("Could not connect to server. %v", err)
	}
	defer cc.Close()
	c := blogpb.NewBlogServiceClient(cc)

	//Create Blog
	blog := &blogpb.Blog{
		AuthorId: "Akhil",
		Title:    "My First Blog",
		Content:  "Content of ther blog",
	}
	id := createBlog(c, blog)

	//Read blog error
	//readBlog(c, "60f50e50db60a7737b8b7c44")
	//Read blog success
	//readBlog(c, id)
	newBlog := &blogpb.Blog{
		Id:       id,
		AuthorId: "Akhil Changed",
		Title:    "My First Blog (updated) ",
		Content:  "Content of ther blog (updated)",
	}

	updateBlog(c, id, newBlog)
	readBlog(c, id)
	deleteBlog(c, id)
	readBlog(c, id)
	listBlog(c)
}

func createBlog(c blogpb.BlogServiceClient, data *blogpb.Blog) string {
	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{
		Blog: data,
	})
	if err != nil {
		log.Fatalf("Error while creating blog")
	}
	fmt.Printf("Blog created : %v\n", res)
	return res.Blog.GetId()
}

func readBlog(c blogpb.BlogServiceClient, id string) {

	res, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{
		Id: id,
	})
	if err != nil {
		fmt.Printf("Error while reading blog %v\n", err)
		return
	}

	fmt.Printf("BlogData: %v\n", res.Blog)
}

func updateBlog(c blogpb.BlogServiceClient, id string, blog *blogpb.Blog) {

	res, err := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{
		Blog: blog,
	})
	if err != nil {
		fmt.Printf("Error while updating blog %v\n", err)
		return
	}

	fmt.Printf("Updated Blog: %v\n", res.GetBlog())
}

func deleteBlog(c blogpb.BlogServiceClient, id string) {

	_, err := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{
		BlogId: id,
	})

	if err != nil {
		fmt.Printf("Falied to delete blog %v\n", err)
		return
	}

	fmt.Println("Deleted successfully")
}

func listBlog(c blogpb.BlogServiceClient) {

	stream, err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		fmt.Printf("error while opening stream %v\n", err)
		return
	}

	for {
		res, e := stream.Recv()
		if e == io.EOF {
			break
		}
		if e != nil {
			fmt.Printf("error while processing stream response %v\n", err)
			break
		}
		fmt.Printf("Blog Data: %v\n", res.GetBlog())
	}
	stream.CloseSend()
}
