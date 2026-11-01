package auth

import (
	"context"
	"errors"
	"fmt"
	"go-ddd-book-manager/internal/config"
	dc "go-ddd-book-manager/internal/domain/common"
	"net/http"
	"net/url"
	"time"
	"uuid"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
)

type claims struct {
	Type  string `json:"typ"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Realm struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

func (*claims) Validate(context.Context) error { return nil }

type jwtVerifier struct{ v *validator.Validator }

func NewVerifier(i do.Injector) (TokenVerifier, error) {
	config := do.MustInvoke[config.Config](i)
	return newVerifier(config.OIDCIssuer, config.OIDCAudience, 5*time.Minute)
}

func newVerifier(issuer, audience string, cacheTTL time.Duration) (TokenVerifier, error) {
	u, err := url.Parse(issuer)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || audience == "" {
		return nil, errors.New("invalid OIDC issuer or audience")
	}

	p := jwks.NewCachingProvider(
		u,
		cacheTTL,
		jwks.WithCustomClient(&http.Client{Timeout: 5 * time.Second}),
		jwks.WithSynchronousRefresh(true),
	)
	key := func(ctx context.Context) (any, error) {
		k, e := p.KeyFunc(ctx)
		if e != nil {
			return nil, fmt.Errorf("%w: %v", ErrProvider, e)
		}
		return k, nil
	}
	v, err := validator.New(
		key,
		validator.RS256,
		issuer,
		[]string{audience},
		validator.WithCustomClaims(func() validator.CustomClaims { return &claims{} }),
	)
	if err != nil {
		return nil, err
	}

	return &jwtVerifier{v}, nil
}

func (v *jwtVerifier) Verify(ctx context.Context, token string) (Identity, error) {
	result, err := v.v.ValidateToken(ctx, token)
	if err != nil {
		if errors.Is(err, ErrProvider) {
			return Identity{}, err
		}
		return Identity{}, ErrUnauthorized
	}

	c, ok := result.(*validator.ValidatedClaims)
	if !ok {
		return Identity{}, errors.New("invalid validator result")
	}

	custom, ok := c.CustomClaims.(*claims)
	if !ok {
		return Identity{}, errors.New("invalid validator claims")
	}

	id, err := uuid.Parse(c.RegisteredClaims.Subject)
	if err != nil || id == uuid.Nil() || c.RegisteredClaims.Expiry == 0 || custom.Type != "Bearer" {
		return Identity{}, ErrUnauthorized
	}

	hasAdmin := lo.Contains(custom.Realm.Roles, "Admin")
	role := lo.Ternary(hasAdmin, dc.AdminRole, dc.UserRole)

	return Identity{ID: id, Name: custom.Name, Email: custom.Email, Role: role}, nil
}
