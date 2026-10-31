package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	dc "go-ddd-book-manager/internal/domain/common"
	jose "gopkg.in/go-jose/go-jose.v2"
	"gopkg.in/go-jose/go-jose.v2/jwt"
)

func TestVerifier(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	other, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	var current atomic.Pointer[rsa.PrivateKey]
	current.Store(key)
	var requests atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]string{"jwks_uri": server.URL + "/jwks"})
		case "/jwks":
			requests.Add(1)
			_ = json.NewEncoder(w).
				Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &current.Load().PublicKey, KeyID: "key", Algorithm: "RS256", Use: "sig"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	v, err := newVerifier(server.URL, "api", 30*time.Millisecond)
	require.NoError(t, err)
	token := func(k *rsa.PrivateKey, issuer, audience, typ string, expiry int64, extra map[string]interface{}) string {
		s, e := jose.NewSigner(
			jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: k, KeyID: "key"}},
			nil,
		)
		require.NoError(t, e)
		m := map[string]interface{}{
			"iss":   issuer,
			"sub":   "00000000-0000-4000-8000-000000000001",
			"aud":   audience,
			"typ":   typ,
			"name":  "Name",
			"email": "test@example.com",
		}
		if expiry != 0 {
			m["exp"] = expiry
		}

		for k, v := range extra {
			m[k] = v
		}
		out, e := jwt.Signed(s).Claims(m).CompactSerialize()
		require.NoError(t, e)
		return out
	}
	future := time.Now().Add(time.Hour).Unix()
	valid := token(key, server.URL, "api", "Bearer", future, nil)
	id, err := v.Verify(context.Background(), valid)
	require.NoError(t, err)
	require.Equal(t, dc.UserRole, id.Role)
	id, err = v.Verify(
		context.Background(),
		token(
			key,
			server.URL,
			"api",
			"Bearer",
			future,
			map[string]interface{}{
				"realm_access": map[string]interface{}{
					"roles": []string{"User", "Admin", "System"},
				},
			},
		),
	)
	require.NoError(t, err)
	require.Equal(t, dc.AdminRole, id.Role)
	require.EqualValues(t, 1, requests.Load())
	for _, s := range []string{token(other, server.URL, "api", "Bearer", future, nil), token(key, "wrong", "api", "Bearer", future, nil), token(key, server.URL, "other", "Bearer", future, nil), token(key, server.URL, "api", "ID", future, nil), token(key, server.URL, "api", "Bearer", time.Now().Add(-time.Hour).Unix(), nil), token(key, server.URL, "api", "Bearer", 0, nil), token(key, server.URL, "api", "Bearer", future, map[string]interface{}{"sub": "invalid"})} {
		_, err = v.Verify(context.Background(), s)
		require.ErrorIs(t, err, ErrUnauthorized)
	}
	current.Store(other)
	time.Sleep(40 * time.Millisecond)
	_, err = v.Verify(context.Background(), token(other, server.URL, "api", "Bearer", future, nil))
	require.NoError(t, err)
	require.GreaterOrEqual(t, requests.Load(), int32(2))
}
