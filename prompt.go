// Copyright 2019 Patrick Pacher. All rights reserved. Use of
// this source code is governed by the included Simplified BSD license.

package keyring

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
)

const (
	promptMethodPrompt    = PromptInterface + ".Prompt"
	promptMethodDismiss   = PromptInterface + ".Dismiss"
	promptSignalCompleted = PromptInterface + ".Completed"
)

// Prompt provides interaction with the Prompt interface from Freedesktop.org's Secret Service API
// it's defined at https://specifications.freedesktop.org/secret-service/re05.html
type Prompt interface {
	// Path returns the ObjectPath of the prompt
	Path() dbus.ObjectPath

	// Prompt performs the prompt
	Prompt(windowID string) (<-chan *dbus.Variant, error)

	// Dismiss dismisses the prompt. It is no longer valid after calling Dismiss()
	Dismiss() error
}

// GetPrompt returns a Prompt client for the given path
func GetPrompt(conn *dbus.Conn, path dbus.ObjectPath) Prompt {
	obj := conn.Object(SecretServiceDest, path)

	return &prompt{
		obj:  obj,
		conn: conn,
		path: path,
	}
}

// prompt implements the Prompt interface
type prompt struct {
	conn *dbus.Conn
	path dbus.ObjectPath
	obj  dbus.BusObject
}

// Path returns the ObjectPath of the prompt
func (p *prompt) Path() dbus.ObjectPath {
	return p.path
}

// Prompt performs the prompt
func (p *prompt) Prompt(windowID string) (<-chan *dbus.Variant, error) {
	call := p.obj.AddMatchSignal(PromptInterface, "Completed")
	if call.Err != nil {
		return nil, call.Err
	}

	ch := make(chan *dbus.Variant, 1)

	sig := make(chan *dbus.Signal, 1)
	p.conn.Signal(sig)

	go func() {
		defer close(sig)
		defer p.conn.RemoveSignal(sig)

		var res []interface{}

		for s := range sig {
			fmt.Println(s.Path)
			if s.Path == p.path {
				res = s.Body
				break
			}
		}

		var dismissed bool
		var result dbus.Variant
		if err := dbus.Store(res, &dismissed, &result); err != nil {
			// how to handle that?
			ch <- nil
			log.Println(err.Error())
			return
		}

		if dismissed {
			ch <- nil
			return
		}

		ch <- &result
	}()

	if res := p.obj.Call(promptMethodPrompt, 0, windowID); res.Err != nil {
		return nil, res.Err
	}

	return ch, nil
}

// Dismiss dismisses the prompt. It is no longer valid after calling Dismiss()
func (p *prompt) Dismiss() error {
	return p.obj.Call(promptMethodDismiss, 0).Err
}
