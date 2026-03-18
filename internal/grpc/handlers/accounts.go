package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DigitLock/expense-tracker/internal/grpc/interceptors"
	pb "github.com/DigitLock/expense-tracker/internal/grpc/pb"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

type AccountHandler struct {
	pb.UnimplementedAccountServiceServer
	repos *repository.Repositories
}

func NewAccountHandler(repos *repository.Repositories) *AccountHandler {
	return &AccountHandler{repos: repos}
}

func (h *AccountHandler) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	accounts, err := h.repos.Accounts.ListByFamily(ctx, familyID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list accounts: %v", err)
	}

	pbAccounts := make([]*pb.Account, 0, len(accounts))
	for _, a := range accounts {
		balance, _, err := h.repos.Accounts.GetBalance(ctx, a.ID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
		}

		balanceFloat, _ := balance.Float64()

		pbAccounts = append(pbAccounts, &pb.Account{
			Id:          a.ID.String(),
			Name:        a.Name,
			Type:        a.Type,
			Currency:    a.Currency,
			Balance:     balanceFloat,
			Description: a.Description.String,
			IsActive:    a.IsActive,
			CreatedAt:   timestamppb.New(a.CreatedAt),
			UpdatedAt:   timestamppb.New(a.UpdatedAt),
		})
	}

	return &pb.ListAccountsResponse{
		Accounts: pbAccounts,
		Total:    int32(len(pbAccounts)),
	}, nil
}
