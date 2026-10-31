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

func Restore(audit common.EntityAudit, name, email string, role common.ActorRole) User {
	return User{
		audit: audit,
		name:  RestoreName(name),
		email: RestoreEmail(email),
		role:  role,
	}
}

func New(
	ctx common.DomainContext,
	id uuid.UUID,
	name Name,
	email Email,
	role common.ActorRole,
) (User, error) {
	if role == common.SystemRole {
		return User{}, RoleInvalidErr
	}
	audit, err := common.NewEntityAuditWithID(ctx, userCretePermission{}, id)
	if err != nil {
		return User{}, err
	}
	return User{audit, name, email, role}, nil
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
func (u *User) Sync(ctx common.DomainContext, name Name, email Email, role common.ActorRole) error {
	actor := ctx.Actor()
	permission := userPermission{actor: actor, userID: u.ID()}
	if u.name == name && u.email == email && u.role == role {
		return nil
	}
	if err := u.audit.Update(ctx, permission); err != nil {
		return err
	}
	u.name, u.email, u.role = name, email, role
	return nil
}
