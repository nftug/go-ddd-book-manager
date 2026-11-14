package command_test

import (
	"context"
	"errors"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/application/user/command"
	dto "go-ddd-book-manager/internal/application/user/dto"
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
		require.Equal(t, id, got.ID())
		repo.AssertExpectations(t)
	})
	t.Run("concurrent creation is reloaded and synchronized", func(t *testing.T) {
		repo := &userRepositoryMock{}
		user := existingUser(t, id)
		changed := request
		changed.Name = "New Name"
		changed.Role = dc.AdminRole
		mock.InOrder(
			repo.On("FindByID", ctx, id).Return(nil, nil).Once(),
			repo.On("Create", ctx, mock.Anything).Return(dc.ErrConflict).Once(),
			repo.On("FindByID", ctx, id).Return(&user, nil).Once(),
			repo.On("Update", ctx, mock.MatchedBy(func(u du.User) bool {
				return u.Name() == changed.Name && u.Role() == changed.Role
			})).Return(nil).Once(),
		)
		actor, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, changed)
		require.NoError(t, err)
		require.Equal(t, dc.NewActor(id, changed.Name, changed.Role), actor)
		repo.AssertExpectations(t)
	})
	t.Run("creation conflict without a matching user remains a conflict", func(t *testing.T) {
		repo := &userRepositoryMock{}
		mock.InOrder(
			repo.On("FindByID", ctx, id).Return(nil, nil).Once(),
			repo.On("Create", ctx, mock.Anything).Return(dc.ErrConflict).Once(),
			repo.On("FindByID", ctx, id).Return(nil, nil).Once(),
		)
		_, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, request)
		require.ErrorIs(t, err, dc.ErrConflict)
		repo.AssertExpectations(t)
	})
	t.Run("creation errors other than conflicts are preserved", func(t *testing.T) {
		repo := &userRepositoryMock{}
		failure := errors.New("create failed")
		repo.On("FindByID", ctx, id).Return(nil, nil).Once()
		repo.On("Create", ctx, mock.Anything).Return(failure).Once()
		_, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, request)
		require.ErrorIs(t, err, failure)
		repo.AssertExpectations(t)
	})
	t.Run("unchanged user is not saved again", func(t *testing.T) {
		repo := &userRepositoryMock{}
		user := existingUser(t, id)
		repo.On("FindByID", ctx, id).Return(&user, nil).Once()
		_, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, request)
		require.NoError(t, err)
		require.True(t, user.Audit().Updated().IsNone())
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
		require.Equal(t, id, got.ID())
		repo.AssertExpectations(t)
	})
	t.Run(
		"update conflict succeeds when the reloaded user is already synchronized",
		func(t *testing.T) {
			repo := &userRepositoryMock{}
			stale := existingUser(t, id)
			synchronized := existingUser(t, id)
			changed := request
			changed.Name = "New Name"
			mock.InOrder(
				repo.On("FindByID", ctx, id).Return(&stale, nil).Once(),
				repo.On("Update", ctx, mock.Anything).Return(dc.ErrConflict).Once(),
				repo.On("FindByID", ctx, id).Return(&synchronized, nil).Once(),
			)
			// The reloaded User must already reflect the requested profile.
			syncProfile(t, &synchronized, changed)
			actor, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, changed)
			require.NoError(t, err)
			require.Equal(t, dc.NewActor(id, changed.Name, changed.Role), actor)
			// The initial update conflicts; no further update is needed after reload.
			repo.AssertNumberOfCalls(t, "Update", 1)
			repo.AssertExpectations(t)
		},
	)
	t.Run("update conflict resynchronizes and retries update once", func(t *testing.T) {
		repo := &userRepositoryMock{}
		stale := existingUser(t, id)
		reloaded := existingUser(t, id)
		changed := request
		changed.Name = "New Name"
		mock.InOrder(
			repo.On("FindByID", ctx, id).Return(&stale, nil).Once(),
			repo.On("Update", ctx, mock.Anything).Return(dc.ErrConflict).Once(),
			repo.On("FindByID", ctx, id).Return(&reloaded, nil).Once(),
			repo.On("Update", ctx, mock.MatchedBy(func(u du.User) bool { return u.Name() == changed.Name })).
				Return(nil).
				Once(),
		)
		actor, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, changed)
		require.NoError(t, err)
		require.Equal(t, dc.NewActor(id, changed.Name, changed.Role), actor)
		repo.AssertNumberOfCalls(t, "Update", 2)
		repo.AssertExpectations(t)
	})
	t.Run("update conflict returns reload error", func(t *testing.T) {
		repo := &userRepositoryMock{}
		stale := existingUser(t, id)
		failure := errors.New("reload failed")
		mock.InOrder(
			repo.On("FindByID", ctx, id).Return(&stale, nil).Once(),
			repo.On("Update", ctx, mock.Anything).Return(dc.ErrConflict).Once(),
			repo.On("FindByID", ctx, id).Return(nil, failure).Once(),
		)
		_, err := newUserCommandService(
			t,
			repo,
			command.NewSyncUser,
		).Execute(ctx, dto.SyncUserRequest{ID: id, Name: "New Name", Email: request.Email, Role: request.Role})
		require.ErrorIs(t, err, failure)
		repo.AssertExpectations(t)
	})
	t.Run("update conflict without a reloaded user remains a conflict", func(t *testing.T) {
		repo := &userRepositoryMock{}
		stale := existingUser(t, id)
		mock.InOrder(
			repo.On("FindByID", ctx, id).Return(&stale, nil).Once(),
			repo.On("Update", ctx, mock.Anything).Return(dc.ErrConflict).Once(),
			repo.On("FindByID", ctx, id).Return(nil, nil).Once(),
		)
		changed := request
		changed.Name = "New Name"
		_, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, changed)
		require.ErrorIs(t, err, dc.ErrConflict)
		repo.AssertExpectations(t)
	})
	t.Run("second update conflict is returned", func(t *testing.T) {
		repo := &userRepositoryMock{}
		stale := existingUser(t, id)
		reloaded := existingUser(t, id)
		changed := request
		changed.Name = "New Name"
		mock.InOrder(
			repo.On("FindByID", ctx, id).Return(&stale, nil).Once(),
			repo.On("Update", ctx, mock.Anything).Return(dc.ErrConflict).Once(),
			repo.On("FindByID", ctx, id).Return(&reloaded, nil).Once(),
			repo.On("Update", ctx, mock.Anything).Return(dc.ErrConflict).Once(),
		)
		_, err := newUserCommandService(t, repo, command.NewSyncUser).Execute(ctx, changed)
		require.ErrorIs(t, err, dc.ErrConflict)
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

func syncProfile(t *testing.T, user *du.User, req dto.SyncUserRequest) {
	t.Helper()
	_, err := user.Sync(
		dc.NewDomainContext(dc.NewSystemActor(), fixedIDs{}, fixedClock{}),
		req.Name,
		req.Email,
		req.Role,
	)
	require.NoError(t, err)
}

func newUserCommandService[T any](
	t *testing.T,
	repo du.Repository,
	constructor func(do.Injector) (T, error),
) T {
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
