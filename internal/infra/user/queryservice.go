package user

import (
	"context"
	"database/sql"
	"errors"
	dto "go-ddd-book-manager/internal/application/user/dto"
	aq "go-ddd-book-manager/internal/application/user/query"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/uptrace/bun"
)

type queryService struct{ db *bun.DB }

func NewQueryService(i do.Injector) (aq.QueryService, error) {
	return &queryService{do.MustInvoke[*bun.DB](i)}, nil
}

func (s *queryService) FindByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	var m model
	err := s.db.NewSelect().Model(&m).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	response := m.toResponse()
	return &response, nil
}
