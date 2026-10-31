package command

import (
	"context"
	dto "go-ddd-book-manager/internal/application/user/dto"
	dc "go-ddd-book-manager/internal/domain/common"
	du "go-ddd-book-manager/internal/domain/user"

	"github.com/samber/do/v2"
)

type SyncUser interface {
	Execute(
		ctx context.Context,
		req dto.SyncUserRequest,
	) (dto.UserResponse, error)
}

type syncUser struct {
	repo  du.Repository
	clock dc.Clock
	ids   dc.UUIDGenerator
}

func NewSyncUser(i do.Injector) (SyncUser, error) {
	return &syncUser{
		repo:  do.MustInvoke[du.Repository](i),
		clock: do.MustInvoke[dc.Clock](i),
		ids:   do.MustInvoke[dc.UUIDGenerator](i),
	}, nil
}

func (s *syncUser) Execute(
	ctx context.Context,
	req dto.SyncUserRequest,
) (dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return dto.UserResponse{}, err
	}
	dctx := dc.NewDomainContext(dc.NewSystemActor(), s.ids, s.clock)

	if user == nil {
		u, err := du.New(dctx, req.ID, req.Name, req.Email, req.Role)
		if err != nil {
			return dto.UserResponse{}, err
		}
		if err = s.repo.Create(ctx, u); err != nil {
			return dto.UserResponse{}, err
		}
		return dto.FromDomain(u), nil
	}

	changed, err := user.Sync(dctx, req.Name, req.Email, req.Role)
	if err != nil {
		return dto.UserResponse{}, err
	}
	if changed {
		if err = s.repo.Update(ctx, *user); err != nil {
			return dto.UserResponse{}, err
		}
	}
	return dto.FromDomain(*user), nil
}
