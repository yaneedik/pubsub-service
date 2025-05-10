package main

import (
	"context"
	
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/yaneedik/pubsub-service/config"
	"github.com/yaneedik/pubsub-service/internal/service"
	"github.com/yaneedik/pubsub-service/internal/subpub"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Config{}
	cfg.GRPC.Port = 50051

	bus := subpub.NewSubPub()
	defer bus.Close(context.Background())

	grpcServer := grpc.NewServer()
	service.RegisterPubSubServer(grpcServer, service.NewPubSubService(bus))

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("Server listening on :%d", cfg.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	grpcServer.GracefulStop()
	log.Println("server stopped")
}

