package handlers

import (
	"context"
	"time"

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
	"github.com/DigitLock/expense-tracker/internal/service/transaction"
)

// TransactionHandler is a thin gRPC adapter over the transaction service.
type TransactionHandler struct {
	pb.UnimplementedTransactionServiceServer
	svc *transaction.Service
}

func NewTransactionHandler(repos *repository.Repositories) *TransactionHandler {
	return &TransactionHandler{svc: transaction.New(repos.Transactions, repos.Accounts, repos.Categories)}
}

func (h *TransactionHandler) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	filter := transaction.ListFilter{
		Type:    req.Type,
		Page:    int(req.GetPage()),
		PerPage: int(req.GetPerPage()),
	}
	if req.AccountId != nil {
		accountID, err := uuid.Parse(req.GetAccountId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid account_id")
		}
		filter.AccountID = &accountID
	}
	if req.Month != nil {
		if _, err := time.Parse("2006-01", req.GetMonth()); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid month format, expected YYYY-MM")
		}
		filter.Month = req.GetMonth()
	}

	result, err := h.svc.List(ctx, familyID, filter)
	if err != nil {
		return nil, errmap.ToStatus(err).Err()
	}

	pbTxs := make([]*pb.Transaction, len(result.Transactions))
	for i, t := range result.Transactions {
		pbTxs[i] = toPbTransaction(t)
	}
	return &pb.ListTransactionsResponse{
		Transactions: pbTxs,
		Page:         int32(result.Page),
		PerPage:      int32(result.PerPage),
		Total:        int32(result.Total),
		TotalPages:   int32(result.TotalPages),
	}, nil
}

func (h *TransactionHandler) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.TransactionResponse, error) {
	familyID, userID, err := identity(ctx)
	if err != nil {
		return nil, err
	}

	tx, serr := h.svc.Create(ctx, familyID, userID, transaction.CreateTransactionInput{
		Type:        req.GetType(),
		Amount:      decimal.NewFromFloat(req.GetAmount()).Round(2), // double -> decimal at the edge
		Currency:    req.GetCurrency(),
		CategoryID:  req.GetCategoryId(),
		AccountID:   req.GetAccountId(),
		Description: req.GetDescription(),
		Date:        req.GetDate(),
	})
	if serr != nil {
		return nil, errmap.ToStatus(serr).Err()
	}
	return &pb.TransactionResponse{Transaction: toPbTransaction(tx)}, nil
}

func (h *TransactionHandler) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.TransactionResponse, error) {
	familyID, userID, err := identity(ctx)
	if err != nil {
		return nil, err
	}
	id, perr := uuid.Parse(req.GetId())
	if perr != nil {
		return nil, errmap.ToStatus(domain.Errorf(domain.ErrInvalidArgument, "invalid transaction id")).Err()
	}

	in := transaction.UpdateTransactionInput{
		Type:        req.Type,
		Currency:    req.Currency,
		CategoryID:  req.CategoryId,
		AccountID:   req.AccountId,
		Description: req.Description,
		Date:        req.Date,
	}
	if req.Amount != nil {
		amt := decimal.NewFromFloat(req.GetAmount()).Round(2)
		in.Amount = &amt
	}

	tx, serr := h.svc.Update(ctx, familyID, userID, id, in)
	if serr != nil {
		return nil, errmap.ToStatus(serr).Err()
	}
	return &pb.TransactionResponse{Transaction: toPbTransaction(tx)}, nil
}

func (h *TransactionHandler) DeleteTransaction(ctx context.Context, req *pb.DeleteTransactionRequest) (*pb.DeleteResponse, error) {
	familyID, userID, err := identity(ctx)
	if err != nil {
		return nil, err
	}
	id, perr := uuid.Parse(req.GetId())
	if perr != nil {
		return nil, errmap.ToStatus(domain.Errorf(domain.ErrInvalidArgument, "invalid transaction id")).Err()
	}

	if serr := h.svc.Delete(ctx, familyID, userID, id); serr != nil {
		return nil, errmap.ToStatus(serr).Err()
	}
	return &pb.DeleteResponse{Success: true, Message: "Transaction deleted successfully"}, nil
}

// identity extracts family and user IDs from the auth context.
func identity(ctx context.Context) (uuid.UUID, uuid.UUID, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return uuid.Nil, uuid.Nil, status.Error(codes.Unauthenticated, "missing family context")
	}
	userID, ok := interceptors.UserIDFromContext(ctx)
	if !ok {
		return uuid.Nil, uuid.Nil, status.Error(codes.Unauthenticated, "missing user context")
	}
	return familyID, userID, nil
}

func toPbTransaction(t domain.Transaction) *pb.Transaction {
	amount, _ := t.Amount.Float64()
	amountBase, _ := t.AmountBase.Float64()
	return &pb.Transaction{
		Id:           t.ID.String(),
		Type:         t.Type,
		Amount:       amount,
		Currency:     t.Currency,
		AmountBase:   amountBase,
		BaseCurrency: t.BaseCurrency,
		AccountId:    t.AccountID.String(),
		CategoryId:   t.CategoryID.String(),
		CategoryName: "", // not part of the domain; left empty (matches prior behavior)
		Description:  t.Description,
		Date:         t.Date.Format("2006-01-02"),
		CreatedAt:    timestamppb.New(t.CreatedAt),
		UpdatedAt:    timestamppb.New(t.UpdatedAt),
	}
}
