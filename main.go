package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"github.com/Prototype-1/freelanceX_proposal_service/config"
	"github.com/Prototype-1/freelanceX_proposal_service/internal/handler"
	"github.com/Prototype-1/freelanceX_proposal_service/internal/repository"
	"github.com/Prototype-1/freelanceX_proposal_service/internal/service"
	"github.com/Prototype-1/freelanceX_proposal_service/proto"
	"google.golang.org/grpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.LoadConfig()

	client, err := mongo.Connect(nil, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(nil)

	proposalRepo := repository.NewProposalRepository(client)
	proposalService := service.NewProposalService(proposalRepo)
	proposalHandler := handler.NewProposalHandler(proposalService)

	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Run every 5 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("Checking and expiring proposals...")
				if err := proposalRepo.ExpireProposals(nil); err != nil {
					log.Printf("Error expiring proposals: %v", err)
				}
			}
		}
	}()

	lis, err := net.Listen("tcp", cfg.ServerPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.ServerPort, err)
	}

	grpcServer := grpc.NewServer()
	proposal.RegisterProposalServiceServer(grpcServer, proposalHandler)

	fmt.Printf("Starting gRPC server on port %s...\n", cfg.ServerPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
