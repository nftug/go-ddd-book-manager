package query

import (
	"context"
	ac "go-ddd-book-manager/internal/application/common"
	dto "go-ddd-book-manager/internal/application/user/dto"
	dc "go-ddd-book-manager/internal/domain/common"

	"github.com/samber/do/v2"
)

type GetUserDetails struct{ query QueryService }

func NewGetUserDetails(i do.Injector) (*GetUserDetails, error) {
	return &GetUserDetails{query: do.MustInvoke[QueryService](i)}, nil
}

func (s *GetUserDetails) Execute(ctx context.Context, actor dc.Actor) (dto.UserResponse, error) {
	u, err := ac.FindRequired(func() (*dto.UserResponse, error) {
		return s.query.FindByID(ctx, actor.ID())
	})
	if err != nil {
		return dto.UserResponse{}, err
	}
	return *u, nil
}
