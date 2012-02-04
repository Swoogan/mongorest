package main

import (
	"http"
	"log"
	"launchpad.net/mgo"
	"bitbucket.org/Swoogan/mongorest"
)

func main() {
	log.Printf("Connecting to mongodb")

	session, err := mgo.Mongo("localhost")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer session.Close()

	db := session.DB("test")

	mongorest.New(db, "customers")
	mongorest.New(db, "employees")

	log.Printf("About to listen on 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
