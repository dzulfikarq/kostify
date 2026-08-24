package application

import (
	"context"

	"github.com/dzulfikarq/kostify/backend/internal/domain"
)

type WishlistUsecase struct {
	repo     domain.WishlistRepository
	props    domain.PropertyRepository
}

func NewWishlistUsecase(repo domain.WishlistRepository, props domain.PropertyRepository) *WishlistUsecase {
	return &WishlistUsecase{repo: repo, props: props}
}

func (uc *WishlistUsecase) Add(ctx context.Context, userID, propertyID string) error {
	p, err := uc.props.FindByID(ctx, propertyID)
	if err != nil {
		return err
	}
	if p.Status != domain.PropertyPublished {
		return domain.Conflict("Hanya kost yang sudah tayang dapat ditambahkan ke wishlist")
	}
	return uc.repo.Add(ctx, userID, propertyID)
}

func (uc *WishlistUsecase) Remove(ctx context.Context, userID, propertyID string) error {
	return uc.repo.Remove(ctx, userID, propertyID)
}

func (uc *WishlistUsecase) List(ctx context.Context, userID string, f domain.ListParams) ([]domain.PropertyWithStats, int64, error) {
	return uc.repo.List(ctx, userID, f)
}

type DashboardUsecase struct {
	repo domain.DashboardRepository
}

func NewDashboardUsecase(repo domain.DashboardRepository) *DashboardUsecase {
	return &DashboardUsecase{repo: repo}
}

func (uc *DashboardUsecase) Get(ctx context.Context, userID string, role domain.Role) (map[string]any, error) {
	switch role {
	case domain.RoleOwner:
		return uc.repo.OwnerStats(ctx, userID)
	case domain.RoleSuperAdmin:
		return uc.repo.AdminStats(ctx)
	default:
		return uc.repo.TenantStats(ctx, userID)
	}
}
