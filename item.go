// Copyright 2019 Patrick Pacher. All rights reserved. Use of
// this source code is governed by the included Simplified BSD license.

package keyring

import (
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	// Methods
	itemMethodDelete    = ItemInterface + ".Delete"
	itemMethodGetSecret = ItemInterface + ".GetSecret"
	itemMethodSetSecret = ItemInterface + ".SetSecret"

	// Properties
	itemPropLocked     = ItemInterface + ".Locked"
	itemPropAttributes = ItemInterface + ".Attributes"
	itemPropLabel      = ItemInterface + ".Label"
	itemPropCreated    = ItemInterface + ".Created"
	itemPropModified   = ItemInterface + ".Modified"
)

// Item implements a wrapper for org.freedesktop.Secret.Item as defined
// here https://specifications.freedesktop.org/secret-service/re03.html
type Item interface {
	// Locked returns true if the item is currently locked
	Locked() (bool, error)

	// Unlock unlocks the item and handles any prompt that might be required
	Unlock() (bool, error)

	// GetAttributes returns the items attributes
	GetAttributes() (map[string]string, error)

	// SetAttributes sets the items attributes
	SetAttributes(map[string]string) error

	// GetLabel returns the label of the item
	GetLabel() (string, error)

	// SetLabel sets the item's label
	SetLabel(string) error

	// Delete deletes the item any handles any prompt that might be required
	Delete() error

	// GetSecret returns the secret of the item
	GetSecret(session dbus.ObjectPath) (*Secret, error)

	// SetSecret sets the secret of the item
	SetSecret(dbus.ObjectPath, []byte, string) error

	// GetCreated returns the time the item has been created
	GetCreated() (time.Time, error)

	// GetModified returns the time the item has been last modified
	GetModified() (time.Time, error)
}

// GetItem returns a new item client for the specified path
func GetItem(conn *dbus.Conn, path dbus.ObjectPath) (Item, error) {
	obj := conn.Object(SecretServiceDest, path)
	i := &item{
		path: path,
		conn: conn,
		obj:  obj,
	}

	if _, err := i.GetLabel(); err != nil {
		return nil, err
	}

	return i, nil
}

// item implements the Item interface
type item struct {
	path dbus.ObjectPath
	conn *dbus.Conn
	obj  dbus.BusObject
}

// Locked returns true if the item is currently locked
func (i *item) Locked() (bool, error) {
	v, err := i.obj.GetProperty(itemPropLocked)
	if err != nil {
		return false, err
	}

	if b, ok := v.Value().(bool); ok {
		return b, nil
	}

	return false, ErrInvalidType("bool", v.Value())
}

// Unlock unlocks the item and handles any prompt that might be required
func (i *item) Unlock() (bool, error) {
	service, err := GetSecretService(i.conn)
	if err != nil {
		return false, err
	}

	if _, err := service.Unlock([]dbus.ObjectPath{i.path}); err != nil {
		return false, err
	}

	return true, nil
}

// GetAttributes returns the items attributes
func (i *item) GetAttributes() (map[string]string, error) {
	v, err := i.obj.GetProperty(itemPropAttributes)
	if err != nil {
		return nil, err
	}

	if b, ok := v.Value().(map[string]string); ok {
		return b, nil
	}

	return nil, ErrInvalidType("map[string]string", v.Value())
}

// SetAttributes sets the items attributes
func (i *item) SetAttributes(m map[string]string) error {
	return i.obj.SetProperty(itemPropAttributes, m)
}

// GetLabel returns the label of the item
func (i *item) GetLabel() (string, error) {
	v, err := i.obj.GetProperty(itemPropLabel)
	if err != nil {
		return "", err
	}

	if s, ok := v.Value().(string); ok {
		return s, nil
	}

	return "", ErrInvalidType("string", v.Value())
}

// SetLabel sets the item's label
func (i *item) SetLabel(l string) error {
	return i.obj.SetProperty(itemPropLabel, l)
}

// Delete deletes the item any handles any prompt that might be required
func (i *item) Delete() error {
	call := i.obj.Call(itemMethodDelete, 0)
	if call.Err != nil {
		return call.Err
	}

	var prompt dbus.ObjectPath
	if err := call.Store(&prompt); err != nil {
		return err
	}

	if prompt != "/" {
		p := GetPrompt(i.conn, prompt)
		res, err := p.Prompt("")
		if err != nil {
			return err
		}

		result := <-res
		if result == nil {
			return fmt.Errorf("prompt dismissed")
		}
	}

	return nil
}

// GetSecret returns the secret of the item
func (i *item) GetSecret(session dbus.ObjectPath) (*Secret, error) {
	var s Secret

	call := i.obj.Call(itemMethodGetSecret, 0, session)
	if call.Err != nil {
		return nil, call.Err
	}

	if err := call.Store(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// SetSecret sets the secret of the item
func (i *item) SetSecret(session dbus.ObjectPath, secret []byte, contentType string) error {
	call := i.obj.Call(itemMethodSetSecret, 0, Secret{
		ContentType: contentType,
		Value:       secret,
		Parameters:  []byte(""),
		Session:     session,
	})
	return call.Err
}

// GetCreated returns the time the item has been created
func (i *item) GetCreated() (time.Time, error) {
	v, err := i.obj.GetProperty(itemPropCreated)
	if err != nil {
		return time.Time{}, err
	}

	u, ok := v.Value().(uint64)
	if !ok {
		return time.Time{}, ErrInvalidType("uint64", v.Value())
	}

	return time.Unix(int64(u), 0), nil
}

// GetModified returns the time the item has been last modified
func (i *item) GetModified() (time.Time, error) {
	v, err := i.obj.GetProperty(itemPropModified)
	if err != nil {
		return time.Time{}, err
	}

	u, ok := v.Value().(uint64)
	if !ok {
		return time.Time{}, ErrInvalidType("uint64", v.Value())
	}

	return time.Unix(int64(u), 0), nil
}
