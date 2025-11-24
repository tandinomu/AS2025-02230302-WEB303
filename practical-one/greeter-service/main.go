package main

import (
	"context"
	"fmt"
	"log"
	"net"
	pb "practical-one/proto/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	pb.UnimplementedGreeterServiceServer
	timeClient pb.TimeServiceClient // Client to call the time-service
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("Received SayHello request for: %v", in.GetName())
	timeReq := &pb.TimeRequest{}
	timeRes, err := s.timeClient.GetTime(ctx, timeReq)
	if err != nil {
		log.Printf("Failed to call time-service: %v", err)
		return nil, err
	}
	message := fmt.Sprintf("Hello %s! The current time is %s", in.GetName(), timeRes.GetCurrentTime())
	return &pb.HelloResponse{Message: message}, nil
}

func main() {
	// Address 'time-service:50052' matches the service name in docker-compose.yml
	conn, err := grpc.Dial("time-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to time-service: %v", err)
	}
	defer conn.Close()
	timeClient := pb.NewTimeServiceClient(conn)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServiceServer(s, &server{timeClient: timeClient})
	log.Printf("Greeter service listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}