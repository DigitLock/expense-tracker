package handlers

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"

	"github.com/DigitLock/expense-tracker/internal/grpc/interceptors"
	pb "github.com/DigitLock/expense-tracker/internal/grpc/pb"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

type TransactionHandler struct {
	pb.UnimplementedTransactionServiceServer
	repos *repository.Repositories
}

func NewTransactionHandler(repos *repository.Repositories) *TransactionHandler {
	return &TransactionHandler{repos: repos}
}

func (h *TransactionHandler) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	filter := repository.TransactionFilter{
		FamilyID: familyID,
		Limit:    perPage,
		Offset:   offset,
	}

	if req.Type != nil {
		filter.Type = req.Type
	}

	if req.AccountId != nil {
		accountID, err := uuid.Parse(*req.AccountId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid account_id: %v", err)
		}
		filter.AccountID = &accountID
	}

	if req.Month != nil {
		start, end, err := parseMonth(*req.Month)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid month format, expected YYYY-MM: %v", err)
		}
		filter.StartDate = &start
		filter.EndDate = &end
	}

	transactions, total, err := h.repos.Transactions.ListFiltered(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list transactions: %v", err)
	}

	pbTransactions := make([]*pb.Transaction, 0, len(transactions))
	for _, t := range transactions {
		amount, _ := t.Amount.Float64()
		amountBase, _ := t.AmountBase.Float64()

		pbTransactions = append(pbTransactions, &pb.Transaction{
			Id:           t.ID.String(),
			Type:         t.Type,
			Amount:       amount,
			Currency:     t.Currency,
			AmountBase:   amountBase,
			BaseCurrency: "RSD",
			AccountId:    t.AccountID.String(),
			CategoryId:   t.CategoryID.String(),
			CategoryName: "",
			Description:  t.Description.String,
			Date:         t.TransactionDate.Time.Format("2006-01-02"),
			CreatedAt:    timestamppb.New(t.CreatedAt),
			UpdatedAt:    timestamppb.New(t.UpdatedAt),
		})
	}

	totalPages := int32(0)
	if total > 0 {
		totalPages = int32((total + int64(perPage) - 1) / int64(perPage))
	}

	return &pb.ListTransactionsResponse{
		Transactions: pbTransactions,
		Page:         page,
		PerPage:      perPage,
		Total:        int32(total),
		TotalPages:   totalPages,
	}, nil
}

func parseMonth(month string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("expected YYYY-MM format: %w", err)
	}
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	return start, end, nil
}
