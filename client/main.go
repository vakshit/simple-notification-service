package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	pb "github.com/vakshit-zomato/assignment-protos"
	"google.golang.org/grpc"
)

func main() {
	timePtr := flag.String("time", "", "Notification time in format YYYY-MM-DD HH:MM")
	messagePtr := flag.String("message", "", "Notification message")
	emailPtr := flag.String("email", "", "User email")

	flag.Parse()

	if *timePtr == "" || *messagePtr == "" || *emailPtr == "" {
		log.Fatalf("time, message, and email are required parameters")
		os.Exit(1)
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewNotificationServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.CreateNotification(ctx, &pb.CreateNotificationRequest{Time: *timePtr, Message: *messagePtr, Email: *emailPtr})
	if err != nil {
		log.Fatalf("could not create notification: %v", err)
	}

	log.Printf("Notification ID: %s, Status: %s", r.GetId(), r.GetStatus())
}
