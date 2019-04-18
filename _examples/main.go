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

	// session is required to create/retrieve secrets
	session, err := svc.OpenSession()
	checkErr(err)
	fmt.Println("session-path: " + session.Path())

	defer func() { checkErr(session.Close()) }()

	// Get all collections available
	collection, err := svc.GetAllCollections()
	checkErr(err)
	var testColl keyring.Collection

	for _, c := range collection {
		l, err := c.GetLabel()
		checkErr(err)
		fmt.Println(c.Path(), " => ", l)
		if l == "test" {
			testColl = c
		}
	}

	// either create a collection or remove it
	if testColl == nil {
		col, err := svc.CreateCollection("test", "")
		checkErr(err)

		item, err := col.CreateItem(session.Path(), "test-item", map[string]string{"application": "test"}, []byte("my-key"), "text/plain", false)
		checkErr(err)

		l, err := item.GetLabel()
		checkErr(err)

		fmt.Println("new-item: ", l)
	} else {
		fmt.Println("current items")
		items, err := testColl.GetAllItems()
		checkErr(err)

		for _, i := range items {
			l, err := i.GetLabel()
			checkErr(err)
			fmt.Println("item: ", l)
		}

		fmt.Println("deleting collection")
		checkErr(testColl.Delete())
	}
}
