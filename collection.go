package keyring

import (
	"fmt"

	"github.com/godbus/dbus"
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

	// Lock locks the collection
	Lock() error

	// Delete deletes the collection
	Delete() error

	// GetAllItems returns all items in the collection
	GetAllItems() ([]dbus.ObjectPath, error)

	// SearchItems searches for items in the collection
	SearchItems(attrs map[string]string) ([]dbus.ObjectPath, error)

	// CreateItem creates a new item inside the collection optionally overwritting an
	// existing one
	CreateItem(session dbus.ObjectPath, label string, attr map[string]string, secret []byte, contentType string, replace bool) (dbus.ObjectPath, error)
}

type collection struct {
	path dbus.ObjectPath
	obj  dbus.BusObject
}

// GetCollection returns a collection object for the specified path
func GetCollection(conn *dbus.Conn, path dbus.ObjectPath) (Collection, error) {
	obj := conn.Object(SecretServiceDest, dbus.ObjectPath(path))
	coll := &collection{
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
	v, err := c.obj.GetProperty(CollectionInterface + ".Label")
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
	return c.obj.SetProperty(CollectionInterface+".Lable", l)
}

// Locked returns true if the collection is locked
func (c *collection) Locked() (bool, error) {
	v, err := c.obj.GetProperty(CollectionInterface + ".Locked")
	if err != nil {
		return false, err
	}

	if b, ok := v.Value().(bool); ok {
		return b, nil
	}

	return false, ErrInvalidType("bool", v.Value())
}

// Lock locks the collection
func (c *collection) Lock() error {
	return fmt.Errorf("not yet implemented")
}

// Delete deletes the collection
func (c *collection) Delete() error {
	return fmt.Errorf("not yet implemented")
}

// GetAllItems returns all items in the collection
func (c *collection) GetAllItems() ([]dbus.ObjectPath, error) {
	v, err := c.obj.GetProperty(CollectionInterface + ".Items")
	if err != nil {
		return nil, err
	}

	if s, ok := v.Value().([]dbus.ObjectPath); ok {
		return s, nil
	}

	return nil, ErrInvalidType("[]string", v.Value())
}

// SearchItems searches for items in the collection
func (c *collection) SearchItems(attrs map[string]string) ([]dbus.ObjectPath, error) {
	call := c.obj.Call(CollectionInterface+".SearchItems", 0, attrs)

	if call.Err != nil {
		fmt.Println(call.Err.Error())
		return nil, call.Err
	}

	list, ok := call.Body[0].([]dbus.ObjectPath)
	if !ok {
		return nil, ErrInvalidType("[]string", call.Body[0])
	}

	return list, nil
}

// CreateItem creates a new item inside the collection optionally overwritting an
// existing one
func (c *collection) CreateItem(session dbus.ObjectPath, label string, attr map[string]string, secret []byte, contentType string, replace bool) (dbus.ObjectPath, error) {
	sec := Secret{
		Session:     session,
		Parameters:  []byte(""),
		Value:       secret,
		ContentType: contentType,
	}

	call := c.obj.Call(CollectionInterface+".CreateItem", 0, map[string]dbus.Variant{
		SecretServicePrefix + "Item.Label":      dbus.MakeVariant(label),
		SecretServicePrefix + "Item.Attributes": dbus.MakeVariant(attr),
	}, sec, replace)

	if call.Err != nil {
		return "", call.Err
	}

	if len(call.Body) != 2 {
		return "", fmt.Errorf("expected 2 results but got %d", len(call.Body))
	}

	itemPath, ok := call.Body[0].(dbus.ObjectPath)
	if !ok {
		return "", ErrInvalidType("ObjectPath", call.Body[0])
	}

	return itemPath, nil
}
