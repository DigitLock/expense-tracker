package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/DigitLock/expense-tracker/internal/auth"
	grpchandlers "github.com/DigitLock/expense-tracker/internal/grpc/handlers"
	"github.com/DigitLock/expense-tracker/internal/grpc/interceptors"
	pb "github.com/DigitLock/expense-tracker/internal/grpc/pb"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	port       int
}

func NewServer(repos *repository.Repositories, jwtService *auth.JWTService, port int) *Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingInterceptor(),
			interceptors.AuthInterceptor(jwtService),
		),
	)

	pb.RegisterAccountServiceServer(grpcServer, grpchandlers.NewAccountHandler(repos))
	pb.RegisterTransactionServiceServer(grpcServer, grpchandlers.NewTransactionHandler(repos))
	pb.RegisterCategoryServiceServer(grpcServer, grpchandlers.NewCategoryHandler(repos))
	pb.RegisterReportServiceServer(grpcServer, grpchandlers.NewReportHandler(repos))
	pb.RegisterAuthServiceServer(grpcServer, grpchandlers.NewAuthHandler(repos, jwtService))

	reflection.Register(grpcServer)
	
	return &Server{
		grpcServer: grpcServer,
		port:       port,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
