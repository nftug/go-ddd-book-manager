package user

import (
	"go-ddd-book-manager/internal/domain/common"
	"uuid"
)

type User struct {
	audit common.EntityAudit
	name  Name
	email Email
	role  common.ActorRole
}

func Restore(audit common.EntityAudit, name, email string, role common.ActorRole) (User, error) {
	if err := validateRole(role); err != nil {
		return User{}, err
	}
	return User{
		audit: audit,
		name:  RestoreName(name),
		email: RestoreEmail(email),
		role:  role,
	}, nil
}

func New(
	ctx common.DomainContext,
	id uuid.UUID,
	name string,
	email string,
	role common.ActorRole,
) (User, error) {
	if err := validateRole(role); err != nil {
		return User{}, err
	}
	validatedName, err := NewName(name)
	if err != nil {
		return User{}, err
	}
	validatedEmail, err := NewEmail(email)
	if err != nil {
		return User{}, err
	}
	audit, err := common.NewEntityAuditWithID(ctx, userSyncPermission{actor: ctx.Actor()}, id)
	if err != nil {
		return User{}, err
	}
	return User{audit, validatedName, validatedEmail, role}, nil
}

func (u User) Audit() common.EntityAudit { return u.audit }
func (u User) ID() uuid.UUID             { return u.audit.ID() }
func (u User) Name() string              { return u.name.Value() }
func (u User) Email() string             { return u.email.Value() }
func (u User) Role() common.ActorRole    { return u.role }
func (u User) Reference() Reference      { return NewReference(u.ID(), u.name) }
func (u User) Actor() common.Actor {
	return common.NewActor(u.ID(), u.name.Value(), u.role)
}

// Sync applies the authoritative identity-provider profile. Unchanged profiles do not update audit data.
func (u *User) Sync(
	ctx common.DomainContext,
	name string,
	email string,
	role common.ActorRole,
) (bool, error) {
	permission := userSyncPermission{actor: ctx.Actor()}
	if err := common.EnsureCanUpdate(permission); err != nil {
		return false, err
	}

	if err := validateRole(role); err != nil {
		return false, err
	}
	validatedName, err := NewName(name)
	if err != nil {
		return false, err
	}
	validatedEmail, err := NewEmail(email)
	if err != nil {
		return false, err
	}
	if u.name == validatedName && u.email == validatedEmail && u.role == role {
		return false, nil
	}
	if err := u.audit.Update(ctx, permission); err != nil {
		return false, err
	}
	u.name, u.email, u.role = validatedName, validatedEmail, role
	return true, nil
}

func (u User) ValidateDeletion(actor common.Actor) error {
	permission := userSyncPermission{actor: actor}
	return common.EnsureCanDelete(permission)
}

func validateRole(role common.ActorRole) error {
	if !role.IsValid() || role == common.SystemRole {
		return RoleInvalidErr
	}
	return nil
}
