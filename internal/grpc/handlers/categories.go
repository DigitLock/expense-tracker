package handlers

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/grpc/errmap"
	"github.com/DigitLock/expense-tracker/internal/grpc/interceptors"
	pb "github.com/DigitLock/expense-tracker/internal/grpc/pb"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/category"
)

// CategoryHandler is a thin gRPC adapter over the category service.
type CategoryHandler struct {
	pb.UnimplementedCategoryServiceServer
	svc *category.Service
}

func NewCategoryHandler(repos *repository.Repositories) *CategoryHandler {
	return &CategoryHandler{svc: category.New(repos.Categories)}
}

func (h *CategoryHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	// parent_id filter is not part of the service surface (mirrors REST); type
	// and include_inactive are honored.
	cats, err := h.svc.List(ctx, familyID, req.GetType(), req.GetIncludeInactive())
	if err != nil {
		return nil, errmap.ToStatus(err).Err()
	}

	pbCats := make([]*pb.Category, len(cats))
	for i, c := range cats {
		pbCats[i] = toPbCategory(c)
	}
	return &pb.ListCategoriesResponse{Categories: pbCats, Total: int32(len(pbCats))}, nil
}

func (h *CategoryHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CategoryResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}

	in := category.CreateCategoryInput{Name: req.GetName(), Type: req.GetType()}
	if pid := req.GetParentId(); pid != "" {
		in.ParentID = &pid
	}

	cat, err := h.svc.Create(ctx, familyID, in)
	if err != nil {
		return nil, errmap.ToStatus(err).Err()
	}
	return &pb.CategoryResponse{Category: toPbCategory(cat)}, nil
}

func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.CategoryResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, errmap.ToStatus(domain.Errorf(domain.ErrInvalidArgument, "invalid category id")).Err()
	}

	// proto3 optional fields arrive as pointers — pass straight through.
	cat, serr := h.svc.Update(ctx, familyID, id, category.UpdateCategoryInput{
		Name:     req.Name,
		ParentID: req.ParentId,
		IsActive: req.IsActive,
	})
	if serr != nil {
		return nil, errmap.ToStatus(serr).Err()
	}
	return &pb.CategoryResponse{Category: toPbCategory(cat)}, nil
}

func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteResponse, error) {
	familyID, ok := interceptors.FamilyIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing family context")
	}
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, errmap.ToStatus(domain.Errorf(domain.ErrInvalidArgument, "invalid category id")).Err()
	}

	if serr := h.svc.Delete(ctx, familyID, id); serr != nil {
		return nil, errmap.ToStatus(serr).Err()
	}
	return &pb.DeleteResponse{Success: true, Message: "Category deactivated successfully"}, nil
}

func toPbCategory(c domain.Category) *pb.Category {
	parentID := ""
	if c.ParentID != nil {
		parentID = c.ParentID.String()
	}
	return &pb.Category{
		Id:        c.ID.String(),
		Name:      c.Name,
		Type:      c.Type,
		ParentId:  parentID,
		IsActive:  c.IsActive,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}
