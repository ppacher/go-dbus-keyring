package keyring

import (
	"github.com/godbus/dbus"
)

type Session interface {
	Path() dbus.ObjectPath
	Close() error
}

func NewSession(conn *dbus.Conn, path dbus.ObjectPath) (Session, error) {
	obj := conn.Object(SecretServiceDest, dbus.ObjectPath(path))

	return &session{
		path: path,
		obj:  obj,
	}, nil
}

type session struct {
	path dbus.ObjectPath
	obj  dbus.BusObject
}

func (s *session) Path() dbus.ObjectPath {
	return s.path
}

func (s *session) Close() error {
	return s.obj.Call(SessionInterface+".Close", 0).Err
}
