package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/akhil4chelsia/grpc-go-microservice/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type BlogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
}

type server struct {
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Creating blog.")
	blog := req.GetBlog()
	data := &BlogItem{
		AuthorID: blog.AuthorId,
		Title:    blog.Title,
		Content:  blog.GetContent(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error %v", err),
		)
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal,
			fmt.Sprintf("Cannot convert to OID %v", ok),
		)
	}
	blog.Id = oid.Hex()
	return &blogpb.CreateBlogResponse{
		Blog: blog,
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Reading blog.")
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Unable to parse object id from hex %v", err))
	}
	filter := bson.M{"_id": id}
	res := collection.FindOne(context.Background(), filter)
	data := &BlogItem{}
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Not found blog with id %v", id))
	}

	return &blogpb.ReadBlogResponse{
		Blog: dataToBlog(data),
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("Updating blog request")
	blog := req.GetBlog()
	id, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Unable to parse object id from hex %v\n", err))
	}
	data := &BlogItem{}
	filter := bson.M{"_id": id}
	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Not found blog with id %v\n", id))
	}
	data.AuthorID = blog.AuthorId
	data.Title = blog.Title
	data.Content = blog.Content
	_, updateErr := collection.ReplaceOne(context.Background(), filter, data)
	if updateErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Failed to update blog %v\n", updateErr),
		)
	}
	return &blogpb.UpdateBlogResponse{
		Blog: dataToBlog(data),
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Deleting blog...")
	id, err := primitive.ObjectIDFromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Unable to parse object id from hex %v\n", err))
	}
	filter := bson.M{"_id": id}
	res, delErr := collection.DeleteOne(context.Background(), filter)
	if delErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Failed to delete blog %v\n", delErr),
		)
	}

	if res.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Not found blog with id %v\n", id))
	}

	return &blogpb.DeleteBlogResponse{
		BlogId: id.Hex(),
	}, nil
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("Streaming blog data")
	cur, err := collection.Find(context.Background(), primitive.D{{}})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unexpected error while quering db %v\n", err),
		)
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		data := &BlogItem{}
		e := cur.Decode(data)
		if e != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data from db %v\n", e),
			)
		}
		stream.Send(&blogpb.ListBlogResponse{Blog: dataToBlog(data)})
	}

	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unexpected error while processing data from db %v\n", err),
		)
	}

	return nil
}

func dataToBlog(data *BlogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Title:    data.Title,
		Content:  data.Content,
	}
}

var collection *mongo.Collection

func main() {

	//logs error line number incase of app crash
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//Connect to mongodb
	fmt.Println("Connecting to Mongodb")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Error while connecting to Mongodb %v", err)
	}
	client.Connect(context.TODO())
	collection = client.Database("mydb").Collection("blog")

	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to start listner. %v", err)
	}
	s := grpc.NewServer()
	blogpb.RegisterBlogServiceServer(s, &server{})
	reflection.Register(s)

	go func() {
		fmt.Println("Starting blog server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to start server. %v", err)
		}
	}()
	// Wait for control C to exit
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing listner")
	lis.Close()
	fmt.Println("Closing Mongodb connection")
	client.Disconnect(context.TODO())
	fmt.Println("Server stopped gracefully.")
}
