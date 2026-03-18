package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpchandler "tracker-task/internal/handler/grpc"
	"tracker-task/internal/infra/memory"
	"tracker-task/internal/usecase"

	pb "tracker-task/gen/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	shipmentRepo := memory.NewShipmentRepository()
	eventRepo := memory.NewEventRepository()
	svc := usecase.NewShipmentService(shipmentRepo, eventRepo)
	handler := grpchandler.NewShipmentHandler(svc)

	server := grpc.NewServer()
	pb.RegisterShipmentServiceServer(server, handler)
	reflection.Register(server)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	go func() {
		log.Printf("gRPC server listening on :%s", port)
		if err := server.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gRPC server...")
	server.GracefulStop()
	log.Println("server stopped")
}
