// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package errors

import (
	"context"
	"errors"
	"testing"

	factory "github.com/dharmab/hyperboard/internal/db/factory"
	models "github.com/dharmab/hyperboard/internal/db/models"
	"github.com/stephenafamo/bob"
)

func TestTagCategoryUniqueConstraintErrors(t *testing.T) {
	if testDB == nil {
		t.Skip("No database connection provided")
	}

	f := factory.New()
	tests := []struct {
		name         string
		expectedErr  *UniqueConstraintError
		conflictMods func(context.Context, *testing.T, bob.Executor, *models.TagCategory) factory.TagCategoryModSlice
	}{
		{
			name:        "ErrUniqueTagCategoriesPkey",
			expectedErr: TagCategoryErrors.ErrUniqueTagCategoriesPkey,
			conflictMods: func(ctx context.Context, t *testing.T, exec bob.Executor, obj *models.TagCategory) factory.TagCategoryModSlice {
				shouldUpdate := false
				updateMods := make(factory.TagCategoryModSlice, 0, 1)

				if shouldUpdate {
					if err := obj.Update(ctx, exec, f.NewTagCategoryWithContext(ctx, updateMods...).BuildSetter()); err != nil {
						t.Fatalf("Error updating object: %v", err)
					}
				}

				return factory.TagCategoryModSlice{
					factory.TagCategoryMods.ID(obj.ID),
				}
			},
		},
		{
			name:        "ErrUniqueTagCategoriesNameKey",
			expectedErr: TagCategoryErrors.ErrUniqueTagCategoriesNameKey,
			conflictMods: func(ctx context.Context, t *testing.T, exec bob.Executor, obj *models.TagCategory) factory.TagCategoryModSlice {
				shouldUpdate := false
				updateMods := make(factory.TagCategoryModSlice, 0, 1)

				if shouldUpdate {
					if err := obj.Update(ctx, exec, f.NewTagCategoryWithContext(ctx, updateMods...).BuildSetter()); err != nil {
						t.Fatalf("Error updating object: %v", err)
					}
				}

				return factory.TagCategoryModSlice{
					factory.TagCategoryMods.Name(obj.Name),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			t.Cleanup(cancel)

			tx, err := testDB.Begin(ctx)
			if err != nil {
				t.Fatalf("Couldn't start database transaction: %v", err)
			}

			defer func() {
				if err := tx.Rollback(ctx); err != nil {
					t.Fatalf("Error rolling back transaction: %v", err)
				}
			}()

			var exec bob.Executor = tx

			obj, err := f.NewTagCategoryWithContext(ctx, factory.TagCategoryMods.WithParentsCascading()).Create(ctx, exec)
			if err != nil {
				t.Fatal(err)
			}

			obj2, err := f.NewTagCategoryWithContext(ctx).Create(ctx, exec)
			if err != nil {
				t.Fatal(err)
			}

			err = obj2.Update(ctx, exec, f.NewTagCategoryWithContext(ctx, tt.conflictMods(ctx, t, exec, obj)...).BuildSetter())
			if !errors.Is(ErrUniqueConstraint, err) {
				t.Fatalf("Expected: %s, Got: %v", tt.name, err)
			}
			if !errors.Is(tt.expectedErr, err) {
				t.Fatalf("Expected: %s, Got: %v", tt.expectedErr.Error(), err)
			}
			if !ErrUniqueConstraint.Is(err) {
				t.Fatalf("Expected: %s, Got: %v", tt.name, err)
			}
			if !tt.expectedErr.Is(err) {
				t.Fatalf("Expected: %s, Got: %v", tt.expectedErr.Error(), err)
			}
		})
	}
}
