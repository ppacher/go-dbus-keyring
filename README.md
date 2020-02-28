# go-dbus-keyring

A GoLang module for querying a keyring application implementing the SecretService DBus specification defined [here](https://specifications.freedesktop.org/secret-service/).

It is based on the awesome dbus library [godbus/dbus](https://github.com/godbus/dbus) (which is the only dependecy of the project).

# Features 

- Full SecretService implementation
- Manage collections
- Manage items/secrets
- Automatically handles user prompts

# Missing Features 

- A server package to implement you own keyring manager
- Support for encrypted secrets (currently only PLAIN is supported)
- Support for signals emitted by various SecretService interfaces (only prompts are supported)
- Unit tests :(

# Usage

`go-dbus-keyring` is setup as a go1.12 module and can be added to any project like this:
```bash
go get -u github.com/ppacher/go-dbus-keyring@v1
```

This project follows [Semantic Versioning](https://semver.org/) as required by go-modules. Those, there will be now API changes in major releases!

The documentation for this project is available on [godoc.org](https://godoc.org/github.com/ppacher/go-dbus-keyring). In addition, there's a simple example inside the [_examples](./_examples) directory.

```bash
package main

import (
    "github.com/godbus/dbus/v5"
    keyring "github.com/ppacher/go-dbus-keyring"
)

func main() {
    bus := dbus.SessionBus()
    
    // Get a SecretService client
    secrets, _ := keyring.GetSecretService(bus)

    // Search for the collection with name "my-collection".
    // You can also use secrets.GetDefaultCollection() or secrets.GetAllCollections()
    collection, _ := secrets.GetCollection("my-collection")
    
    // Search for the item with name "my-password"
    item, _ := collection.GetItem("my-password")
    
    // make sure it is unlocked
    // this also handles any prompt that may be required
    _ = item.Unlock()

    secret, _ := item.GetSecret()
    fmt.Println(string(secret.Value))
}

```

# Contributions

Contributions to this project are welcome! Just fork the repository and create a pull request!
If you need help to get started checkout the github documentation on [creating pull requests](https://help.github.com/en/articles/creating-a-pull-request-from-a-fork).

# License

go-dbus-keyring is available under a Simplified BSD License. See LICENSE file for the full text.

