package main

import (
	"context"
	"log"
	"net"
	"strings"

	pb "github.com/LuhTonkaYeat/bookshelf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedAuthorValidatorServer
}

func (s *server) Validate(c context.Context, r *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	author := strings.ToLower(r.Author)

	if author == "rowling" || author == "king" {
		return &pb.ValidateResponse{
			Valid:  false,
			Reason: "author is banned",
		}, nil
	}

	return &pb.ValidateResponse{
		Valid:  true,
		Reason: "ok",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthorValidatorServer(s, &server{})

	reflection.Register(s)

	log.Println("Validator service running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
