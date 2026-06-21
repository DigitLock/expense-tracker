package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/DigitLock/expense-tracker/internal/grpc/errmap"
	"github.com/DigitLock/expense-tracker/internal/grpc/interceptors"
	pb "github.com/DigitLock/expense-tracker/internal/grpc/pb"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/report"
)

// ReportHandler is a thin gRPC adapter over the report query service.
type ReportHandler struct {
	pb.UnimplementedReportServiceServer
	svc *report.Service
}

func NewReportHandler(repos *repository.Repositories) *ReportHandler {
	return &ReportHandler{svc: report.New(repos.Transactions, repos.Categories, repos.ExchangeRates)}
}

func (h *ReportHandler) GetSpendingByCategory(ctx context.Context, req *pb.SpendingByCategoryRequest) (*pb.SpendingByCategoryResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	rep, err := h.svc.SpendingByCategory(ctx, familyID, report.SpendingParams{
		StartDate: req.GetStartDate(),
		EndDate:   req.GetEndDate(),
		Currency:  req.GetCurrency(),
		Type:      req.GetType(),
	})
	if err != nil {
		return nil, errmap.ToStatus(err).Err()
	}

	cats := make([]*pb.CategorySpending, len(rep.Categories))
	for i, c := range rep.Categories {
		total, _ := c.TotalAmount.Float64()
		pct, _ := c.Percentage.Float64()
		avg, _ := c.AveragePerTransaction.Float64()
		cats[i] = &pb.CategorySpending{
			CategoryId:            c.CategoryID.String(),
			CategoryName:          c.CategoryName,
			TotalAmount:           total,
			TransactionCount:      int32(c.TransactionCount),
			Percentage:            pct,
			AveragePerTransaction: avg,
		}
	}
	totalAmount, _ := rep.TotalAmount.Float64()

	return &pb.SpendingByCategoryResponse{
		ReportType:        rep.ReportType,
		Period:            &pb.Period{StartDate: rep.Period.StartDate, EndDate: rep.Period.EndDate},
		Currency:          rep.Currency,
		TransactionType:   rep.TransactionType,
		Categories:        cats,
		TotalAmount:       totalAmount,
		TotalTransactions: int32(rep.TotalTransactions),
		GeneratedAt:       rep.GeneratedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (h *ReportHandler) GetMonthlySummary(ctx context.Context, req *pb.MonthlySummaryRequest) (*pb.MonthlySummaryResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	sum, err := h.svc.MonthlySummary(ctx, familyID, req.GetMonth(), req.GetCurrency())
	if err != nil {
		return nil, errmap.ToStatus(err).Err()
	}

	income, _ := sum.Income.Float64()
	expenses, _ := sum.Expenses.Float64()
	net, _ := sum.NetBalance.Float64()

	return &pb.MonthlySummaryResponse{
		Month:         sum.Month,
		Currency:      sum.Currency,
		TotalIncome:   income,
		TotalExpenses: expenses,
		NetBalance:    net,
		IncomeCount:   int32(sum.IncomeCount),
		ExpenseCount:  int32(sum.ExpenseCount),
		GeneratedAt:   sum.GeneratedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
