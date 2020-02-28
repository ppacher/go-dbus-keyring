// Copyright 2019 Patrick Pacher. All rights reserved. Use of
// this source code is governed by the included Simplified BSD license.

package keyring

import (
	"github.com/godbus/dbus/v5"
)

const (
	sessionMethodClose = SessionInterface + ".Close"
)

// Session allows to interact with the Session interface of Freedesktop.org's Secret Service API
// The session interface is defined at https://specifications.freedesktop.org/secret-service/re01.html
type Session interface {
	// Path returns the object path of the session
	// To get a new session use SecretService.OpenSession()
	Path() dbus.ObjectPath

	// Close closes the session
	Close() error
}

// GetSession returns a new Session for the provided path. Note that session must be opened beforehand
// Use SecretService.OpenSession() to open a new session and return a Session client
func GetSession(conn *dbus.Conn, path dbus.ObjectPath) (Session, error) {
	obj := conn.Object(SecretServiceDest, dbus.ObjectPath(path))

	return &session{
		path: path,
		obj:  obj,
	}, nil
}

// session implements the Session interface
type session struct {
	path dbus.ObjectPath
	obj  dbus.BusObject
}

// Path returns the ObjectPath of the session
func (s *session) Path() dbus.ObjectPath {
	return s.path
}

// Close closes the session
func (s *session) Close() error {
	return s.obj.Call(sessionMethodClose, 0).Err
}
