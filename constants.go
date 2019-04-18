package keyring

import "github.com/godbus/dbus"

const (
	SecretServiceDest   = "org.freedesktop.secrets"
	SecretServicePrefix = "org.freedesktop.Secret."
	SecretServicePath   = "/org/freedesktop/secrets"

	CollectionInterface = SecretServicePrefix + "Collection"
	SessionInterface    = SecretServicePrefix + "Session"
	ServiceInterface    = SecretServicePrefix + "Service"
	DefaultCollection   = SecretServicePath + "/aliases/default"
	SessionCollection   = SecretServicePath + "/collection/session"

	AlgPlain = "plain"
	// AlgDH is not yet supported only AlgPlain is supported
	AlgDH = "dh-ietf1024-sha256-aes128-cbc-pkcs7"
)

// Secret defines the DBUS STRUCT for a
// secret
type Secret struct {
	Session     dbus.ObjectPath
	Parameters  []byte
	Value       []byte
	ContentType string
}
