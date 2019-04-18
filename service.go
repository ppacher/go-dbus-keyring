package keyring

import (
	"fmt"

	"github.com/godbus/dbus"
)

// SecretService manages all the sessions and collections
// it's defined in org.freedesktop.Secret.Service
// https://specifications.freedesktop.org/secret-service/re01.html
type SecretService interface {
	// OpenSession opens a unique session for the calling application
	OpenSession() (Session, error)

	// GetAllCollections returns all collections stored in the secret service
	GetAllCollections() ([]Collection, error)

	// GetDefaultCollection returns the default collection of the secret service
	// ( DBus path = /org/freedesktop/secrets/aliases/default )
	GetDefaultCollection() (Collection, error)

	// SearchItems finds all items in any collection and returns them either
	// in the unlocked or locked slice
	SearchItems(map[string]string) (unlocked []dbus.ObjectPath, locked []dbus.ObjectPath, err error)

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
	call := svc.obj.Call(ServiceInterface+".OpenSession", 0, "plain", dbus.MakeVariant(""))
	if call.Err != nil {
		return nil, call.Err
	}

	if len(call.Body) != 2 {
		return nil, fmt.Errorf("expected 2 results but got %d", len(call.Body))
	}

	path, ok := call.Body[1].(dbus.ObjectPath)
	if ok {
		return NewSession(svc.conn, path)
	}

	return nil, ErrInvalidType("ObjectPath", call.Body[0])
}

// GetAllCollections returns all collections stored in the secret service
func (svc *service) GetAllCollections() ([]Collection, error) {
	v, err := svc.obj.GetProperty(ServiceInterface + ".Collections")
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
func (svc *service) SearchItems(attrs map[string]string) ([]dbus.ObjectPath, []dbus.ObjectPath, error) {
	call := svc.obj.Call(ServiceInterface+".SearchItems", 0, attrs)
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

	return unlocked, locked, nil
}

// GetSecrets returns multiple secrets from different items
func (svc *service) GetSecrets(paths []dbus.ObjectPath, session dbus.ObjectPath) (map[dbus.ObjectPath]*Secret, error) {
	call := svc.obj.Call(ServiceInterface+".GetSecrets", 0, paths, session)
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
	call := svc.obj.Call(ServiceInterface+".ReadAlias", 0, name)
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
	return svc.obj.Call(ServiceInterface+".SetAlias", 0, name, path).Err
}

// RemoveAlias removes the provided alias. This is a utility method for SetAlias(name, "/")
func (svc *service) RemoveAlias(name string) error {
	return svc.SetAlias(name, "/")
}
