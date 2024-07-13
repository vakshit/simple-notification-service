package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/vakshit-zomato/assignment-protos"
	"github.com/vakshit-zomato/assignment-server/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

const (
	mongoURI       = "mongodb://localhost:27017"
	redisAddr      = "localhost:6379"
	redisDB        = 0
	dbName         = "notificationdb"
	collectionName = "notifications"
)

type server struct {
	pb.UnimplementedNotificationServiceServer
	mongoClient *mongo.Client
}

func (s *server) CreateNotification(ctx context.Context, req *pb.CreateNotificationRequest) (*pb.CreateNotificationResponse, error) {
	notif := domain.Notification{Time: req.Time, Message: req.Message, Email: req.Email}
	collection := s.mongoClient.Database(dbName).Collection(collectionName)
	_, err := collection.InsertOne(ctx, notif)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNotificationResponse{Status: "success"}, nil
}

func (s *server) startNotificationChecker() {
	for {
		now := time.Now().Format("2006-01-02 15:04")
		collection := s.mongoClient.Database(dbName).Collection(collectionName)

		// Define the filter for the documents you want to find
		filter := bson.D{{Key: "time", Value: now}}

		// Find all documents matching the filter
		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			log.Fatalf("Failed to find documents: %v", err)
		}
		defer cur.Close(context.TODO())

		// Loop over the cursor to get all documents
		for cur.Next(context.TODO()) {
			var result domain.Notification
			err := cur.Decode(&result)
			if err != nil {
				log.Fatalf("Failed to decode document: %v", err)
			}
			// Process the document
			sendMail(result)
		}

		cur.Close(context.TODO())

		if err := cur.Err(); err != nil {
			log.Fatalf("Cursor error: %v", err)
		}
		time.Sleep(time.Minute)
	}
}

func sendMail(notif domain.Notification) {
	fmt.Println("Sending mail to", notif.Email, "at: ", time.Now().Format("2006-01-02 15:04:05"))
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	srv := &server{mongoClient: mongoClient}

	go srv.startNotificationChecker()

	pb.RegisterNotificationServiceServer(s, srv)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
