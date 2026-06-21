package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/grpc/errmap"
	"github.com/DigitLock/expense-tracker/internal/grpc/interceptors"
	pb "github.com/DigitLock/expense-tracker/internal/grpc/pb"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/account"
)

// AccountHandler is a thin gRPC adapter over the account application service.
type AccountHandler struct {
	pb.UnimplementedAccountServiceServer
	svc *account.Service
}

func NewAccountHandler(repos *repository.Repositories) *AccountHandler {
	return &AccountHandler{svc: account.New(repos.Accounts)}
}

func (h *AccountHandler) ListAccounts(ctx context.Context, _ *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	accounts, err := h.svc.List(ctx, familyID, false)
	if err != nil {
		return nil, errmap.ToStatus(err).Err()
	}

	pbAccounts := make([]*pb.Account, len(accounts))
	for i, a := range accounts {
		pbAccounts[i] = toPbAccount(a)
	}
	return &pb.ListAccountsResponse{Accounts: pbAccounts, Total: int32(len(pbAccounts))}, nil
}

func (h *AccountHandler) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.AccountResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	acc, err := h.svc.Create(ctx, familyID, account.CreateAccountInput{
		Name:     req.GetName(),
		Type:     req.GetType(),
		Currency: req.GetCurrency(),
		// proto Balance is double; normalize float noise to 2dp at the gRPC edge.
		InitialBalance: decimal.NewFromFloat(req.GetInitialBalance()).Round(2),
		Description:    req.GetDescription(),
	})
	if err != nil {
		return nil, errmap.ToStatus(err).Err()
	}
	return &pb.AccountResponse{Account: toPbAccount(acc)}, nil
}

func (h *AccountHandler) UpdateAccount(ctx context.Context, req *pb.UpdateAccountRequest) (*pb.AccountResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, errmap.ToStatus(domain.Errorf(domain.ErrInvalidArgument, "invalid account id")).Err()
	}

	// proto3 optional fields arrive as pointers — pass them straight through.
	acc, err := h.svc.Update(ctx, familyID, id, account.UpdateAccountInput{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	})
	if err != nil {
		return nil, errmap.ToStatus(err).Err()
	}
	return &pb.AccountResponse{Account: toPbAccount(acc)}, nil
}

func (h *AccountHandler) DeleteAccount(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, errmap.ToStatus(domain.Errorf(domain.ErrInvalidArgument, "invalid account id")).Err()
	}

	if err := h.svc.Delete(ctx, familyID, id); err != nil {
		return nil, errmap.ToStatus(err).Err()
	}
	return &pb.DeleteResponse{Success: true, Message: "Account deactivated successfully"}, nil
}

func toPbAccount(a domain.Account) *pb.Account {
	balance, _ := a.CurrentBalance.Float64() // proto Balance is double
	return &pb.Account{
		Id:          a.ID.String(),
		Name:        a.Name,
		Type:        a.Type,
		Currency:    a.Currency,
		Balance:     balance,
		Description: a.Description,
		IsActive:    a.IsActive,
		CreatedAt:   timestamppb.New(a.CreatedAt),
		UpdatedAt:   timestamppb.New(a.UpdatedAt),
	}
}
