package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/spf13/viper"
	pb "github.com/vakshit-zomato/assignment-protos"
	"github.com/vakshit-zomato/assignment-server/domain"
	"github.com/vakshit-zomato/assignment-server/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

const (
	mongoURI       = "mongodb://localhost:27017"
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
	res, err := collection.InsertOne(ctx, notif)
	if err != nil {
		return nil, err
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("failed to convert inserted ID to ObjectID")
	}
	return &pb.CreateNotificationResponse{Id: oid.Hex(), Status: "success"}, nil
}

func (s *server) startNotificationChecker() {
	var cur *mongo.Cursor
	defer cur.Close(context.TODO())
	for {
		now := time.Now().Format("2006-01-02 15:04")
		log.Println("Checking for notifications at: ", now)
		collection := s.mongoClient.Database(dbName).Collection(collectionName)

		// Define the filter for the documents you want to find
		filter := bson.D{{Key: "time", Value: now}}

		// Find all documents matching the filter
		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			log.Fatalf("Failed to find documents: %v", err)
		}

		// Loop over the cursor to get all documents
		var ids []primitive.ObjectID
		var notifs []domain.Notification
		for cur.Next(context.TODO()) {
			var result domain.Notification
			err := cur.Decode(&result)
			if err != nil {
				log.Fatalf("Failed to decode document: %v", err)
			}
			// Process the document
			notifs = append(notifs, result)
			ids = append(ids, result.ID)
		}
		go service.SendEmailBulk(notifs)
		if len(ids) != 0 {
			_, err = collection.DeleteMany(context.TODO(), bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}})
			if err != nil {
				log.Fatalf("Failed to delete documents: %v", err)
			}
			log.Println("Sent notification count: ", len(ids))
		}

		cur.Close(context.TODO())

		if err := cur.Err(); err != nil {
			log.Fatalf("Cursor error: %v", err)
		}
		time.Sleep(time.Minute)
	}
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}

func main() {
	initConfig()
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
