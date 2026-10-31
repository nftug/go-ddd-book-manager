package option_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/common/option"
)

func TestSomeAndNone(t *testing.T) {
	some := option.Some(42)

	assert.True(t, some.IsSome())
	assert.False(t, some.IsNone())

	got, ok := some.Get()
	assert.True(t, ok)
	assert.Equal(t, 42, got)

	none := option.None[int]()

	assert.False(t, none.IsSome())
	assert.True(t, none.IsNone())

	got, ok = none.Get()
	assert.False(t, ok)
	assert.Zero(t, got)
}

func TestNewOptionAndValue(t *testing.T) {
	assert.True(t, option.New[int](nil).IsNone())

	zero := 0
	someZero := option.New(&zero)

	got, ok := someZero.Get()
	assert.True(t, ok)
	assert.Zero(t, got)

	value := 42
	some := option.New(&value)
	value = 100

	gotValue := some.Value()
	require.NotNil(t, gotValue)
	assert.Equal(t, 42, *gotValue)
	*some.Value() = 7

	assert.Equal(t, 42, some.ValueOr(-1))
	assert.Nil(t, option.None[int]().Value())
}

func TestValueOr(t *testing.T) {
	assert.Equal(t, "present", option.Some("present").ValueOr("fallback"))

	assert.Equal(t, "fallback", option.None[string]().ValueOr("fallback"))
}

func TestMapTransformsPresentValue(t *testing.T) {
	got := option.Some(42).Map(func(value int) string {
		assert.Equal(t, 42, value)
		return ""
	})

	assert.True(t, got.IsSome())
	assert.Equal(t, "", got.ValueOr("missing"))
}

func TestMapPreservesAbsenceWithoutCallingTransform(t *testing.T) {
	got := option.None[int]().Map(func(value int) string {
		assert.Fail(t, "transform must not be called for None")
		return "unexpected"
	})

	assert.True(t, got.IsNone())
}

func TestMarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
		want  string
	}{
		{name: "none", value: option.None[int](), want: "null"},
		{name: "some zero", value: option.Some(0), want: "0"},
		{name: "some string", value: option.Some("value"), want: `"value"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.value)
			require.NoError(t, err)

			assert.Equal(t, tc.want, string(got))
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var some option.Option[int]
	require.NoError(t, json.Unmarshal([]byte("42"), &some))

	assert.True(t, some.IsSome())
	assert.Equal(t, 42, some.ValueOr(0))

	require.NoError(t, json.Unmarshal([]byte("null"), &some))

	assert.True(t, some.IsNone())

	require.NoError(t, some.UnmarshalJSON(nil))

	assert.True(t, some.IsNone())
}

func TestUnmarshalJSONPreservesValueOnError(t *testing.T) {
	value := option.Some(42)

	assert.Error(t, json.Unmarshal([]byte(`"not an integer"`), &value))
	assert.Equal(t, option.Some(42), value)
}
