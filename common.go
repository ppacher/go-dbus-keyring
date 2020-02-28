// Copyright 2019 Patrick Pacher. All rights reserved. Use of
// this source code is governed by the included Simplified BSD license.

package keyring

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	SecretServiceDest   = "org.freedesktop.secrets"
	SecretServicePrefix = "org.freedesktop.Secret."
	SecretServicePath   = "/org/freedesktop/secrets"

	CollectionInterface = SecretServicePrefix + "Collection"
	SessionInterface    = SecretServicePrefix + "Session"
	ItemInterface       = SecretServicePrefix + "Item"
	ServiceInterface    = SecretServicePrefix + "Service"
	PromptInterface     = SecretServicePrefix + "Prompt"
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

func ErrInvalidType(expected string, value interface{}) error {
	return fmt.Errorf("invalid type: expected a '%s' but got '%T'", expected, value)
}
