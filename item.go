package keyring

import (
	"errors"
	"fmt"
	"time"

	"github.com/godbus/dbus"
)

// Item implements a wrapper for org.freedesktop.Secret.Item as defined
// here https://specifications.freedesktop.org/secret-service/re03.html
type Item interface {
	IsLocked() (bool, error)
	Unlock() (bool, error)
	GetAttributes() (map[string]string, error)
	SetAttributes(map[string]string) error
	GetLabel() (string, error)
	SetLabel(string) error
	Delete() error
	GetSecret() ([]byte, string, error)
	SetSecret([]byte, string) error
	GetCreated() (time.Time, error)
	GetModified() (time.Time, error)
}

func ErrInvalidType(expected string, value interface{}) error {
	return fmt.Errorf("invalid type: expected a '%s' but got '%T'", expected, value)
}

// item implements the Item interface
type item struct {
	path string
	obj  dbus.BusObject
}

func (i *item) IsLocked() (bool, error) {
	v, err := i.obj.GetProperty("Locked")
	if err != nil {
		return false, err
	}

	if b, ok := v.Value().(bool); ok {
		return b, nil
	}

	return false, ErrInvalidType("bool", v.Value())
}

func (i *item) Unlock() (bool, error) {
	return false, errors.New("not yet implemented")
}

func (i *item) GetAttributes() (map[string]string, error) {
	v, err := i.obj.GetProperty("Attributes")
	if err != nil {
		return nil, err
	}

	if b, ok := v.Value().(map[string]string); ok {
		return b, nil
	}

	return nil, ErrInvalidType("map[string]string", v.Value())
}

func (i *item) SetAttributes(m map[string]string) error {
	return i.obj.SetProperty("Attributes", m)
}

func (i *item) GetLabel() (string, error) {
	v, err := i.obj.GetProperty("Label")
	if err != nil {
		return "", err
	}

	if s, ok := v.Value().(string); ok {
		return s, nil
	}

	return "", ErrInvalidType("string", v.Value())
}

func (i *item) SetLabel(l string) error {
	return i.obj.SetProperty("Label", l)
}

func (i *item) Delete() error {
	call := i.obj.Call("Delete", 0)
	return call.Err
}
