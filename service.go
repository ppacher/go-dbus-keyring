// Copyright 2019 Patrick Pacher. All rights reserved. Use of
// this source code is governed by the included Simplified BSD license.

package keyring

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	// Methods
	// https://specifications.freedesktop.org/secret-service/re01.html#id479701
	serviceMethodOpenSession      = ServiceInterface + ".OpenSession"
	serviceMethodCreateCollection = ServiceInterface + ".CreateCollection"
	serviceMethodSearchItems      = ServiceInterface + ".SearchItems"
	serviceMethodUnlock           = ServiceInterface + ".Unlock"
	serviceMethodLock             = ServiceInterface + ".Lock"
	serviceMethodGetSecrets       = ServiceInterface + ".GetSecrets"
	serviceMethodReadAlias        = ServiceInterface + ".ReadAlias"
	serviceMethodSetAlias         = ServiceInterface + ".SetAlias"

	// Signals
	// https://specifications.freedesktop.org/secret-service/re01.html#id480380
	serviceSignalCollectionCreated = ServiceInterface + ".CollectionCreated"
	serviceSignalCollectionDeleted = ServiceInterface + ".CollectionDeleted"
	serviceSignalCollectionChanged = ServiceInterface + ".CollectionChanged"

	// Properties
	// https://specifications.freedesktop.org/secret-service/re01.html#id480507
	servicePropCollections = ServiceInterface + ".Collections"
)

// SecretService manages all the sessions and collections
// it's defined in org.freedesktop.Secret.Service
// https://specifications.freedesktop.org/secret-service/re01.html
type SecretService interface {
	// OpenSession opens a unique session for the calling application
	OpenSession() (Session, error)

	// GetCollection returns the collection with the given name
	GetCollection(name string) (Collection, error)

	// GetAllCollections returns all collections stored in the secret service
	GetAllCollections() ([]Collection, error)

	// GetDefaultCollection returns the default collection of the secret service
	// ( DBus path = /org/freedesktop/secrets/aliases/default )
	GetDefaultCollection() (Collection, error)

	// SearchItems finds all items in any collection and returns them either
	// in the unlocked or locked slice
	SearchItems(map[string]string) (unlocked []Item, locked []Item, err error)

	// GetSecrets returns multiple secrets from different items
	GetSecrets(paths []dbus.ObjectPath, session dbus.ObjectPath) (map[dbus.ObjectPath]*Secret, error)

	// ReadAlias resolves the alias (like 'default') to the object path of the
	// referenced collection
	ReadAlias(name string) (dbus.ObjectPath, error)

	// SetAlias creates a new alias for the given collection path
	// Note that if path is "/", the alias will be deleted
	// see https://specifications.freedesktop.org/secret-service/re01.html#org.freedesktop.Secret.Service.SetAlias
	SetAlias(name string, path dbus.ObjectPath) error

	// RemoveAlias removes the provided alias. This is a utility method for SetAlias(name, "/")
	RemoveAlias(name string) error

	// CreateCollection creates a new collection with the given properties and an optional alias (leave empty for no alias)
	// It also handles any prompt that may be required
	CreateCollection(label string, alias string) (Collection, error)

	// Lock locks items or collections and handles any prompt that may be required
	Lock(paths []dbus.ObjectPath) ([]dbus.ObjectPath, error)

	// Unlock unlocks items or collections and handles any prompt that may be required
	Unlock(paths []dbus.ObjectPath) ([]dbus.ObjectPath, error)
}

type service struct {
	obj  dbus.BusObject
	conn *dbus.Conn
}

// GetSecretService returns a client to the SecretService (org.freedesktop.secrets)
// on the provided DBus connection
func GetSecretService(conn *dbus.Conn) (SecretService, error) {
	obj := conn.Object(SecretServiceDest, SecretServicePath)

	svc := &service{
		obj:  obj,
		conn: conn,
	}

	return svc, nil
}

// OpenSession opens a unique session for the calling application
func (svc *service) OpenSession() (Session, error) {
	call := svc.obj.Call(serviceMethodOpenSession, 0, "plain", dbus.MakeVariant(""))
	if call.Err != nil {
		return nil, call.Err
	}

	if len(call.Body) != 2 {
		return nil, fmt.Errorf("expected 2 results but got %d", len(call.Body))
	}

	path, ok := call.Body[1].(dbus.ObjectPath)
	if ok {
		return GetSession(svc.conn, path)
	}

	return nil, ErrInvalidType("ObjectPath", call.Body[0])
}

// GetCollection returns the first collection with the given label
func (svc *service) GetCollection(name string) (Collection, error) {
	all, err := svc.GetAllCollections()
	if err != nil {
		return nil, err
	}

	for _, c := range all {
		l, err := c.GetLabel()
		if err != nil {
			return nil, err
		}

		if l == name {
			return c, nil
		}
	}
	return nil, fmt.Errorf("unknown collection")
}

// GetAllCollections returns all collections stored in the secret service
func (svc *service) GetAllCollections() ([]Collection, error) {
	v, err := svc.obj.GetProperty(servicePropCollections)
	if err != nil {
		return nil, err
	}

	paths, ok := v.Value().([]dbus.ObjectPath)
	if !ok {
		return nil, ErrInvalidType("[]ObjectPath", v.Value())
	}

	col := make([]Collection, len(paths))
	for i, p := range paths {
		var err error
		col[i], err = GetCollection(svc.conn, p)

		if err != nil {
			return nil, err
		}
	}

	return col, nil
}

// GetDefaultCollection returns the default collection of the secret service
// ( DBus path = /org/freedesktop/secrets/aliases/default )
func (svc *service) GetDefaultCollection() (Collection, error) {
	return GetCollection(svc.conn, DefaultCollection)
}

// SearchItems finds all items in any collection and returns them either
// in the unlocked or locked slice
func (svc *service) SearchItems(attrs map[string]string) ([]Item, []Item, error) {
	call := svc.obj.Call(serviceMethodSearchItems, 0, attrs)
	if call.Err != nil {
		return nil, nil, call.Err
	}

	if len(call.Body) != 2 {
		return nil, nil, fmt.Errorf("expected 2 results but got %v", len(call.Body))
	}

	var unlocked []dbus.ObjectPath
	var locked []dbus.ObjectPath

	if err := call.Store(&unlocked, &locked); err != nil {
		return nil, nil, err
	}

	unlockedItems := make([]Item, len(unlocked))
	lockedItems := make([]Item, len(locked))

	for i, u := range unlocked {
		item, err := GetItem(svc.conn, u)
		if err != nil {
			return nil, nil, err
		}
		unlockedItems[i] = item
	}

	for i, u := range locked {
		item, err := GetItem(svc.conn, u)
		if err != nil {
			return nil, nil, err
		}
		lockedItems[i] = item
	}

	return unlockedItems, lockedItems, nil
}

// GetSecrets returns multiple secrets from different items
func (svc *service) GetSecrets(paths []dbus.ObjectPath, session dbus.ObjectPath) (map[dbus.ObjectPath]*Secret, error) {
	call := svc.obj.Call(serviceMethodGetSecrets, 0, paths, session)
	if call.Err != nil {
		return nil, call.Err
	}

	var result map[dbus.ObjectPath][]interface{}

	if err := call.Store(&result); err != nil {
		return nil, err
	}

	secrets := make(map[dbus.ObjectPath]*Secret, len(result))
	for path, res := range result {
		var sec Secret

		if err := dbus.Store([]interface{}{res}, &sec); err != nil {
			return nil, err
		}

		secrets[path] = &sec
	}

	return secrets, nil
}

// ReadAlias resolves the alias (like 'default') to the object path of the
// referenced collection
func (svc *service) ReadAlias(name string) (dbus.ObjectPath, error) {
	call := svc.obj.Call(serviceMethodReadAlias, 0, name)
	if call.Err != nil {
		return "", call.Err
	}

	var path dbus.ObjectPath
	if err := call.Store(&path); err != nil {
		return "", err
	}

	if path == dbus.ObjectPath("/") {
		return path, fmt.Errorf("unknown alias")
	}

	return path, nil
}

// SetAlias creates a new alias for the given collection path
// Note that if path is "/", the alias will be deleted
// see https://specifications.freedesktop.org/secret-service/re01.html#org.freedesktop.Secret.Service.SetAlias
func (svc *service) SetAlias(name string, path dbus.ObjectPath) error {
	return svc.obj.Call(serviceMethodSetAlias, 0, name, path).Err
}

// RemoveAlias removes the provided alias. This is a utility method for SetAlias(name, "/")
func (svc *service) RemoveAlias(name string) error {
	return svc.SetAlias(name, "/")
}

// CreateCollection creates a new collection with the given properties and an optional alias (leave empty for no alias)
// It also handles any prompt that may be required
func (svc *service) CreateCollection(label string, alias string) (Collection, error) {

	properties := map[string]dbus.Variant{}
	properties[collectionPropLabel] = dbus.MakeVariant(label)

	call := svc.obj.Call(serviceMethodCreateCollection, 0, properties, alias)
	if call.Err != nil {
		return nil, call.Err
	}

	var collectionPath dbus.ObjectPath
	var promptPath dbus.ObjectPath

	if err := call.Store(&collectionPath, &promptPath); err != nil {
		return nil, err
	}

	// check if a prompt is required
	if promptPath != "/" {
		// assert(collectionPath == "")

		p := GetPrompt(svc.conn, promptPath)
		res, err := p.Prompt("")
		if err != nil {
			return nil, err
		}

		result := <-res
		if result == nil {
			return nil, fmt.Errorf("prompt dismissed")
		}

		var ok bool
		collectionPath, ok = result.Value().(dbus.ObjectPath)
		if !ok {
			return nil, ErrInvalidType("ObjectPath", result.Value())
		}
	}

	col, err := GetCollection(svc.conn, collectionPath)
	if err != nil {
		return nil, err
	}

	return col, nil
}

// Lock locks items or collections and handles any prompt that may be required
func (svc *service) Lock(paths []dbus.ObjectPath) ([]dbus.ObjectPath, error) {
	call := svc.obj.Call(serviceMethodLock, 0, paths)
	if call.Err != nil {
		return nil, call.Err
	}

	var locked []dbus.ObjectPath
	var prompt dbus.ObjectPath
	if err := call.Store(&locked, &prompt); err != nil {
		return nil, err
	}

	if prompt != "/" {
		p := GetPrompt(svc.conn, prompt)
		res, err := p.Prompt("")
		if err != nil {
			return nil, err
		}

		result := <-res
		if result == nil {
			return locked, fmt.Errorf("prompt dismissed")
		}
	}

	return locked, nil
}

// Unlock unlocks items or collections and handles any prompt that may be required
func (svc *service) Unlock(paths []dbus.ObjectPath) ([]dbus.ObjectPath, error) {
	call := svc.obj.Call(serviceMethodUnlock, 0, paths)
	if call.Err != nil {
		return nil, call.Err
	}

	var locked []dbus.ObjectPath
	var prompt dbus.ObjectPath
	if err := call.Store(&locked, &prompt); err != nil {
		return nil, err
	}

	if prompt != "/" {
		p := GetPrompt(svc.conn, prompt)
		res, err := p.Prompt("")
		if err != nil {
			return nil, err
		}

		result := <-res
		if result == nil {
			return locked, fmt.Errorf("prompt dismissed")
		}
	}

	return locked, nil
}
