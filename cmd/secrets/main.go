package main

import (
	"fmt"
	"log"

	keyring "github.com/ppacher/go-dbus-keyring"

	"github.com/godbus/dbus"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	conn, err := dbus.SessionBus()
	checkErr(err)

	svc, err := keyring.GetSecretService(conn)
	checkErr(err)

	session, err := svc.OpenSession()
	checkErr(err)
	fmt.Println("session-path: " + session.Path())

	defer func() { checkErr(session.Close()) }()

	unlocked, locked, err := svc.SearchItems(map[string]string{})
	checkErr(err)
	fmt.Printf("all-items:\n\t%d locked\n\t%d unlocked\n", len(locked), len(unlocked))

	allCollection, err := svc.GetAllCollections()
	checkErr(err)
	fmt.Println("All-Collections:")
	for _, col := range allCollection {
		l, _ := col.GetLabel()
		fmt.Println("\t" + string(col.Path()) + ": " + l)
	}

	col, err := svc.GetDefaultCollection()
	checkErr(err)

	label, err := col.GetLabel()
	checkErr(err)
	fmt.Println("Label: " + label)

	isLocked, err := col.IsLocked()
	checkErr(err)
	fmt.Printf("Locked: %v\n", isLocked)

	itemsPaths, err := col.GetAllItems()
	checkErr(err)
	fmt.Printf("Items: %d\n", len(itemsPaths))

	itemsPaths, err = col.SearchItems(map[string]string{"foo": "bar"})
	checkErr(err)
	fmt.Printf("SearchItems: %d\n", len(itemsPaths))

	result, err := svc.GetSecrets(itemsPaths, session.Path())
	checkErr(err)
	for path, sec := range result {
		fmt.Printf("%s: %#v\n", path, sec)
	}
}
