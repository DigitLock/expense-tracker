package handlers

import (
	"context"

	jwtauth "github.com/DigitLock/expense-tracker/internal/auth"
	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/grpc/errmap"
	pb "github.com/DigitLock/expense-tracker/internal/grpc/pb"
	"github.com/DigitLock/expense-tracker/internal/repository"
	authsvc "github.com/DigitLock/expense-tracker/internal/service/auth"
)

// AuthHandler is a thin gRPC adapter over the shared auth service. Its RPCs are
// whitelisted in the auth interceptor (token is obtained/validated here).
type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	svc *authsvc.Service
}

func NewAuthHandler(repos *repository.Repositories, jwtService *jwtauth.JWTService) *AuthHandler {
	return &AuthHandler{svc: authsvc.New(repos.Users, jwtService)}
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	res, err := h.svc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, errmap.ToStatus(err).Err()
	}
	return &pb.LoginResponse{
		Token:     res.Token,
		User:      toPbUser(res.User),
		ExpiresIn: int32(res.ExpiresIn),
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	// Token comes from the request body (this RPC is public), not metadata.
	valid, user, err := h.svc.ValidateToken(ctx, req.GetToken())
	if err != nil {
		return nil, errmap.ToStatus(err).Err()
	}
	resp := &pb.ValidateTokenResponse{Valid: valid}
	if valid {
		resp.User = toPbUser(user)
	}
	return resp, nil
}

func toPbUser(u domain.User) *pb.User {
	return &pb.User{
		Id:       u.ID,
		Email:    u.Email,
		Name:     u.Name,
		FamilyId: u.FamilyID.String(),
	}
}
