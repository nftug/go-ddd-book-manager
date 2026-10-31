package command_test

import (
	"context"
	"errors"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/application/user/command"
	dto "go-ddd-book-manager/internal/application/user/dto"
	"go-ddd-book-manager/internal/application/user/service"
	"go-ddd-book-manager/internal/common/option"
	dc "go-ddd-book-manager/internal/domain/common"
	du "go-ddd-book-manager/internal/domain/user"
	"testing"
	"time"
	"uuid"
)

type userRepositoryMock struct{ mock.Mock }

func (m *userRepositoryMock) FindByID(ctx context.Context, id uuid.UUID) (*du.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*du.User), args.Error(1)
}

func (m *userRepositoryMock) Create(ctx context.Context, user du.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *userRepositoryMock) Update(ctx context.Context, user du.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *userRepositoryMock) Delete(ctx context.Context, user du.User) error {
	return m.Called(ctx, user).Error(0)
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

type fixedIDs struct{}

func (fixedIDs) NewUUIDv7() uuid.UUID { return uuid.NewV7() }

func existingUser(t *testing.T, id uuid.UUID) du.User {
	t.Helper()
	audit := dc.RestoreEntityAudit(
		id,
		1,
		dc.RestoreAuditRecord(id, "A", fixedClock{}.Now()),
		option.None[dc.AuditRecord](),
	)
	user, err := du.Restore(audit, "A", "a@example.com", dc.UserRole)
	require.NoError(t, err)
	return user
}

func TestSyncUser(t *testing.T) {
	ctx := context.Background()
	id := uuid.NewV7()
	request := dto.SyncUserRequest{ID: id, Name: "A", Email: "a@example.com", Role: dc.UserRole}
	t.Run("new user is created and returned", func(t *testing.T) {
		repo := &userRepositoryMock{}
		repo.On("FindByID", ctx, id).Return(nil, nil).Once()
		repo.On("Create", ctx, mock.MatchedBy(func(u du.User) bool { return u.ID() == id && u.Email() == request.Email })).
			Return(nil).
			Once()
		got, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, request)
		require.NoError(t, err)
		require.Equal(t, id, got.ID)
		repo.AssertExpectations(t)
	})
	t.Run("unchanged user is not saved again", func(t *testing.T) {
		repo := &userRepositoryMock{}
		user := existingUser(t, id)
		repo.On("FindByID", ctx, id).Return(&user, nil).Once()
		got, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, request)
		require.NoError(t, err)
		require.Nil(t, got.Audit.Updated)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})
	t.Run("email-only change is saved", func(t *testing.T) {
		repo := &userRepositoryMock{}
		user := existingUser(t, id)
		changed := request
		changed.Email = "new@example.com"
		repo.On("FindByID", ctx, id).Return(&user, nil).Once()
		repo.On("Update", ctx, mock.MatchedBy(func(u du.User) bool { return u.Email() == changed.Email })).
			Return(nil).
			Once()
		got, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, changed)
		require.NoError(t, err)
		require.Equal(t, changed.Email, got.Email)
		repo.AssertExpectations(t)
	})
	t.Run("repository errors are preserved", func(t *testing.T) {
		failure := errors.New("storage unavailable")
		repo := &userRepositoryMock{}
		repo.On("FindByID", ctx, id).Return(nil, failure).Once()
		_, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, request)
		require.ErrorIs(t, err, failure)
		repo.AssertExpectations(t)
	})
}

func TestResolveActor(t *testing.T) {
	ctx := context.Background()
	id := uuid.NewV7()
	t.Run("missing user is not found", func(t *testing.T) {
		repo := &userRepositoryMock{}
		repo.On("FindByID", ctx, id).Return(nil, nil).Once()
		_, err := newUserCommandService(t, repo, service.NewResolveActor).Execute(ctx, id)
		require.ErrorIs(t, err, dc.ErrNotFound)
		repo.AssertExpectations(t)
	})
	t.Run("existing user becomes actor without being synchronized", func(t *testing.T) {
		repo := &userRepositoryMock{}
		repo.On("FindByID", ctx, id).
			Return(userPtr(existingUser(t, id)), nil).
			Once()
		actor, err := newUserCommandService(t, repo, service.NewResolveActor).Execute(ctx, id)
		require.NoError(t, err)
		require.Equal(t, id, actor.ID())
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})
}

func userPtr(user du.User) *du.User { return &user }

func newUserCommandService[T any](
	t *testing.T,
	repo du.Repository,
	constructor func(do.Injector) (*T, error),
) *T {
	t.Helper()
	injector := do.New()
	do.ProvideValue[du.Repository](injector, repo)
	do.ProvideValue[dc.Clock](injector, fixedClock{})
	do.ProvideValue[dc.UUIDGenerator](injector, fixedIDs{})
	t.Cleanup(func() { injector.Shutdown() })
	service, err := constructor(injector)
	require.NoError(t, err)
	return service
}
