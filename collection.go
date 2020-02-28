// Copyright 2019 Patrick Pacher. All rights reserved. Use of
// this source code is governed by the included Simplified BSD license.

package keyring

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	// Methods
	collectionMethodDelete      = CollectionInterface + ".Delete"
	collectionMethodSearchItems = CollectionInterface + ".SearchItems"
	collectionMethodCreateItem  = CollectionInterface + ".CreateItem"

	// signals
	collectionSignalItemCreated = CollectionInterface + ".ItemCreated"
	collectionSignalItemDeleted = CollectionInterface + ".ItemDeleted"
	collectionSignalItemChanged = CollectionInterface + ".ItemChanged"

	// Properties
	collectionPropLabel    = CollectionInterface + ".Label"
	collectionPropLocked   = CollectionInterface + ".Locked"
	collectionPropItems    = CollectionInterface + ".Items"
	collectionPropCreated  = CollectionInterface + ".Created"
	collectionPropModified = CollectionInterface + ".Modified"
)

// Collection provides access secret collections from org.freedesktop.secret
// The DBus specification for org.freedesktop.Secret.Collection can be found
// at https://specifications.freedesktop.org/secret-service/re02.html
type Collection interface {
	// Path returns the ObjectPath of the collection
	Path() dbus.ObjectPath

	// GetLabel returns the label of the collection
	GetLabel() (string, error)

	// SetLabel sets the label of the connection
	SetLabel(l string) error

	// Locked returns true if the collection is locked
	Locked() (bool, error)

	// Delete deletes the collection and handles any prompt required
	Delete() error

	// GetAllItems returns all items in the collection
	GetAllItems() ([]Item, error)

	// GetItem returns the first item with the given label
	GetItem(name string) (Item, error)

	// SearchItems searches for items in the collection
	SearchItems(attrs map[string]string) ([]Item, error)

	// CreateItem creates a new item inside the collection optionally overwritting an
	// existing one
	CreateItem(session dbus.ObjectPath, label string, attr map[string]string, secret []byte, contentType string, replace bool) (Item, error)
}

type collection struct {
	conn *dbus.Conn
	path dbus.ObjectPath
	obj  dbus.BusObject
}

// GetCollection returns a collection object for the specified path
func GetCollection(conn *dbus.Conn, path dbus.ObjectPath) (Collection, error) {
	obj := conn.Object(SecretServiceDest, dbus.ObjectPath(path))
	coll := &collection{
		conn: conn,
		obj:  obj,
		path: path,
	}

	if _, err := coll.GetLabel(); err != nil {
		return nil, err
	}

	return coll, nil
}

// Path returns the ObjectPath of the collection
func (c *collection) Path() dbus.ObjectPath {
	return c.path
}

// GetLabel returns the label of the collection
func (c *collection) GetLabel() (string, error) {
	v, err := c.obj.GetProperty(collectionPropLabel)
	if err != nil {
		return "", err
	}

	l, ok := v.Value().(string)
	if !ok {
		return "", ErrInvalidType("string", v.Value())
	}

	return l, nil
}

// SetLabel sets the label of the connection
func (c *collection) SetLabel(l string) error {
	return c.obj.SetProperty(collectionPropLabel, l)
}

// Locked returns true if the collection is locked
func (c *collection) Locked() (bool, error) {
	v, err := c.obj.GetProperty(collectionPropLocked)
	if err != nil {
		return false, err
	}

	if b, ok := v.Value().(bool); ok {
		return b, nil
	}

	return false, ErrInvalidType("bool", v.Value())
}

// Delete deletes the collection and handles any prompt required
func (c *collection) Delete() error {
	call := c.obj.Call(collectionMethodDelete, 0)
	if call.Err != nil {
		return call.Err
	}

	var promptPath dbus.ObjectPath
	if err := call.Store(&promptPath); err != nil {
		return err
	}

	if promptPath != "/" {
		p := GetPrompt(c.conn, promptPath)
		res, err := p.Prompt("")
		if err != nil {
			return err
		}

		result := <-res
		if result == nil {
			return fmt.Errorf("prompted dismissed")
		}
	}

	return nil
}

// GetAllItems returns all items in the collection
func (c *collection) GetAllItems() ([]Item, error) {
	v, err := c.obj.GetProperty(collectionPropItems)
	if err != nil {
		return nil, err
	}

	if list, ok := v.Value().([]dbus.ObjectPath); ok {
		items := make([]Item, len(list))
		for i, it := range list {
			items[i], err = GetItem(c.conn, it)
			if err != nil {
				return nil, err
			}
		}

		return items, nil
	}

	return nil, ErrInvalidType("[]string", v.Value())
}

// GetItem returns the first item with the given name
func (c *collection) GetItem(name string) (Item, error) {
	all, err := c.GetAllItems()
	if err != nil {
		return nil, err
	}

	for _, i := range all {
		l, err := i.GetLabel()
		if err != nil {
			return nil, err
		}

		if l == name {
			return i, nil
		}
	}

	return nil, fmt.Errorf("no such item")
}

// SearchItems searches for items in the collection
func (c *collection) SearchItems(attrs map[string]string) ([]Item, error) {
	call := c.obj.Call(collectionMethodSearchItems, 0, attrs)

	if call.Err != nil {
		fmt.Println(call.Err.Error())
		return nil, call.Err
	}

	list, ok := call.Body[0].([]dbus.ObjectPath)
	if !ok {
		return nil, ErrInvalidType("[]string", call.Body[0])
	}

	var err error

	items := make([]Item, len(list))
	for i, it := range list {
		items[i], err = GetItem(c.conn, it)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

// CreateItem creates a new item inside the collection optionally overwritting an
// existing one
func (c *collection) CreateItem(session dbus.ObjectPath, label string, attr map[string]string, secret []byte, contentType string, replace bool) (Item, error) {
	sec := Secret{
		Session:     session,
		Parameters:  []byte(""),
		Value:       secret,
		ContentType: contentType,
	}

	call := c.obj.Call(collectionMethodCreateItem, 0, map[string]dbus.Variant{
		SecretServicePrefix + "Item.Label":      dbus.MakeVariant(label),
		SecretServicePrefix + "Item.Attributes": dbus.MakeVariant(attr),
	}, sec, replace)

	if call.Err != nil {
		return nil, call.Err
	}

	if len(call.Body) != 2 {
		return nil, fmt.Errorf("expected 2 results but got %d", len(call.Body))
	}

	itemPath, ok := call.Body[0].(dbus.ObjectPath)
	if !ok {
		return nil, ErrInvalidType("ObjectPath", call.Body[0])
	}

	return GetItem(c.conn, itemPath)
}
