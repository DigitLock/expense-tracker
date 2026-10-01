package category_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/service/category"
)

// BE-002: restoring a category whose name is taken by an active category must
// return the same duplicate-name error as Create, not a raw 23505.
func TestRestore_NameCollision(t *testing.T) {
	cases := []struct {
		name       string
		activeName string
	}{
		{"same case", "Food"},
		{"case variant", "FOOD"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := setup(t)
			ctx := context.Background()
			deleted, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if err := f.svc.Delete(ctx, f.familyID, deleted.ID); err != nil {
				t.Fatalf("delete: %v", err)
			}
			if _, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: tc.activeName, Type: "expense"}); err != nil {
				t.Fatalf("create active duplicate: %v", err)
			}

			_, err = f.svc.Restore(ctx, deleted.ID)
			if !errors.Is(err, domain.ErrAlreadyExists) {
				t.Fatalf("restore err = %v, want ErrAlreadyExists", err)
			}
			if err.Error() != "category with this name already exists" {
				t.Errorf("error text = %q, want the Create duplicate-name text", err.Error())
			}

			got, err := f.repos.Categories.GetByIDIncludingInactive(ctx, deleted.ID)
			if err != nil {
				t.Fatalf("get: %v", err)
			}
			if got.IsActive {
				t.Error("category should stay inactive after a failed restore")
			}
		})
	}
}

func TestRestore_Success(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	c, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := f.svc.Delete(ctx, f.familyID, c.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	restored, err := f.svc.Restore(ctx, c.ID)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !restored.IsActive || restored.ID != c.ID {
		t.Errorf("restored = %+v, want active category %s", restored, c.ID)
	}
}

// The REST handler matches this repository error text to return 404, so the
// service must pass it through unchanged.
func TestRestore_AlreadyActivePassesThrough(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	c, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = f.svc.Restore(ctx, c.ID)
	if err == nil || err.Error() != "category not found or already active" {
		t.Errorf("restore active err = %v, want \"category not found or already active\"", err)
	}
	if errors.Is(err, domain.ErrAlreadyExists) {
		t.Error("already-active must not map to ErrAlreadyExists")
	}
}
