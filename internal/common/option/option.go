package option

import (
	"encoding/json"
)

type Option[T comparable] struct {
	value    T
	hasValue bool
}

func Some[T comparable](value T) Option[T] {
	return Option[T]{value: value, hasValue: true}
}

func None[T comparable]() Option[T] {
	return Option[T]{hasValue: false}
}

func NewOption[T comparable](value *T) Option[T] {
	if value == nil {
		return None[T]()
	}
	return Some(*value)
}

func (o Option[T]) IsSome() bool { return o.hasValue }

func (o Option[T]) IsNone() bool { return !o.hasValue }

func (o Option[T]) Get() (T, bool) {
	return o.value, o.hasValue
}

func (o Option[T]) Value() *T {
	if !o.hasValue {
		return nil
	}
	copied := o.value
	return &copied
}

func (o Option[T]) ValueOr(defaultValue T) T {
	if !o.hasValue {
		return defaultValue
	}
	return o.value
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if !o.hasValue {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}

func (o *Option[T]) UnmarshalJSON(data []byte) error {
	// Commit-style decode: only mutate receiver after successful decode.
	// Note: When a struct field is omitted, encoding/json does not call UnmarshalJSON at all.
	if data == nil || string(data) == "null" {
		*o = None[T]()
		return nil
	}

	var decoded T
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	*o = Some(decoded)
	return nil
}
