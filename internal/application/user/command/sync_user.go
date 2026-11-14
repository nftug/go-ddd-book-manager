package command

import (
	"context"
	"errors"
	dto "go-ddd-book-manager/internal/application/user/dto"
	dc "go-ddd-book-manager/internal/domain/common"
	du "go-ddd-book-manager/internal/domain/user"
	"uuid"

	"github.com/samber/do/v2"
)

type SyncUser interface {
	Execute(ctx context.Context, req dto.SyncUserRequest) (dc.Actor, error)
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

func (s *syncUser) Execute(ctx context.Context, req dto.SyncUserRequest) (dc.Actor, error) {
	user, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return dc.Actor{}, err
	}

	dctx := dc.NewDomainContext(dc.NewSystemActor(), s.ids, s.clock)
	if user == nil {
		return s.create(ctx, req, dctx)
	}
	return s.update(ctx, user, req, dctx)
}

func (s *syncUser) create(
	ctx context.Context,
	req dto.SyncUserRequest,
	dctx dc.DomainContext,
) (dc.Actor, error) {
	u, err := du.New(dctx, req.ID, req.Name, req.Email, req.Role)
	if err != nil {
		return dc.Actor{}, err
	}
	if err = s.repo.Create(ctx, u); err != nil {
		// Another request may have created this user after the initial lookup.
		user, err := s.findUserAfterConflict(ctx, req.ID, err)
		if err != nil {
			return dc.Actor{}, err
		}
		return s.update(ctx, user, req, dctx)
	}
	return dc.NewActor(u.ID(), u.Name(), u.Role()), nil
}

func (s *syncUser) update(
	ctx context.Context,
	user *du.User,
	req dto.SyncUserRequest,
	dctx dc.DomainContext,
) (dc.Actor, error) {
	changed, err := user.Sync(dctx, req.Name, req.Email, req.Role)
	if err != nil {
		return dc.Actor{}, err
	}
	if changed {
		if err = s.repo.Update(ctx, *user); err != nil {
			user, err = s.findUserAfterConflict(ctx, req.ID, err)
			if err != nil {
				return dc.Actor{}, err
			}

			changed, err = user.Sync(dctx, req.Name, req.Email, req.Role)
			if err != nil {
				return dc.Actor{}, err
			}
			if changed {
				if err = s.repo.Update(ctx, *user); err != nil {
					return dc.Actor{}, err
				}
			}
		}
	}
	return dc.NewActor(user.ID(), user.Name(), user.Role()), nil
}

func (s *syncUser) findUserAfterConflict(
	ctx context.Context,
	id uuid.UUID,
	conflictErr error,
) (*du.User, error) {
	if !errors.Is(conflictErr, dc.ErrConflict) {
		return nil, conflictErr
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, conflictErr
	}
	return user, nil
}
