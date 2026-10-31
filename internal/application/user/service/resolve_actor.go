package service

import (
	"context"

	"github.com/samber/do/v2"
	ac "go-ddd-book-manager/internal/application/common"
	dc "go-ddd-book-manager/internal/domain/common"
	du "go-ddd-book-manager/internal/domain/user"
	"uuid"
)

type ResolveActor struct{ repo du.Repository }

func NewResolveActor(i do.Injector) (*ResolveActor, error) {
	return &ResolveActor{repo: do.MustInvoke[du.Repository](i)}, nil
}

func (s *ResolveActor) Execute(ctx context.Context, id uuid.UUID) (dc.Actor, error) {
	u, err := ac.FindRequired(func() (*du.User, error) {
		return s.repo.FindByID(ctx, id)
	})
	if err != nil {
		return dc.Actor{}, err
	}
	return u.Actor(), nil
}
