package option_test

import (
	"encoding/json"
	"testing"

	"go-ddd-book-manager/internal/common/option"
)

func TestSomeAndNone(t *testing.T) {
	some := option.Some(42)
	if some.IsNone() || !some.IsSome() {
		t.Fatal("Some should report IsSome and not IsNone")
	}
	if got, ok := some.Get(); !ok || got != 42 {
		t.Fatalf("Some.Get() = (%d, %t), want (42, true)", got, ok)
	}

	none := option.None[int]()
	if none.IsSome() || !none.IsNone() {
		t.Fatal("None should report IsNone and not IsSome")
	}
	if got, ok := none.Get(); ok || got != 0 {
		t.Fatalf("None.Get() = (%d, %t), want (0, false)", got, ok)
	}
}

func TestNewOptionAndValue(t *testing.T) {
	if got := option.New[int](nil); got.IsSome() {
		t.Fatal("NewOption(nil) should return None")
	}

	zero := 0
	someZero := option.New(&zero)
	if got, ok := someZero.Get(); !ok || got != 0 {
		t.Fatalf("NewOption(&0).Get() = (%d, %t), want (0, true)", got, ok)
	}

	value := 42
	some := option.New(&value)
	value = 100
	if got := some.Value(); got == nil || *got != 42 {
		t.Fatalf("NewOption should copy its input; Value() = %v, want pointer to 42", got)
	}
	*some.Value() = 7
	if got, _ := some.Get(); got != 42 {
		t.Fatalf("mutating Value() pointer changed Option value to %d, want 42", got)
	}
	if got := option.None[int]().Value(); got != nil {
		t.Fatalf("None.Value() = %v, want nil", got)
	}
}

func TestValueOr(t *testing.T) {
	if got := option.Some("present").ValueOr("fallback"); got != "present" {
		t.Fatalf("Some.ValueOr() = %q, want present", got)
	}
	if got := option.None[string]().ValueOr("fallback"); got != "fallback" {
		t.Fatalf("None.ValueOr() = %q, want fallback", got)
	}
}

func TestMapTransformsPresentValue(t *testing.T) {
	got := option.Some(42).Map(func(value int) string {
		if value != 42 {
			t.Fatalf("transform got %d, want 42", value)
		}
		return ""
	})
	if value, ok := got.Get(); !ok || value != "" {
		t.Fatalf("Map result = (%q, %t), want an empty string present", value, ok)
	}
}

func TestMapPreservesAbsenceWithoutCallingTransform(t *testing.T) {
	got := option.None[int]().Map(func(value int) string {
		t.Fatal("transform must not be called for None")
		return "unexpected"
	})
	if got.IsSome() {
		t.Fatal("Map of None should return None")
	}
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
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("Marshal() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var some option.Option[int]
	if err := json.Unmarshal([]byte("42"), &some); err != nil {
		t.Fatal(err)
	}
	if got, ok := some.Get(); !ok || got != 42 {
		t.Fatalf("Unmarshal number = (%d, %t), want (42, true)", got, ok)
	}

	if err := json.Unmarshal([]byte("null"), &some); err != nil {
		t.Fatal(err)
	}
	if some.IsSome() {
		t.Fatal("Unmarshal null should set Option to None")
	}

	if err := some.UnmarshalJSON(nil); err != nil {
		t.Fatal(err)
	}
	if some.IsSome() {
		t.Fatal("UnmarshalJSON(nil) should set Option to None")
	}
}

func TestUnmarshalJSONPreservesValueOnError(t *testing.T) {
	value := option.Some(42)
	if err := json.Unmarshal([]byte(`"not an integer"`), &value); err == nil {
		t.Fatal("unmarshalling a string into Option[int] should fail")
	}
	if got, ok := value.Get(); !ok || got != 42 {
		t.Fatalf("failed Unmarshal changed Option to (%d, %t), want (42, true)", got, ok)
	}
}
