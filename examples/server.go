package main

import (
	"http"
	"log"
	"launchpad.net/mgo"
	"bitbucket.com/Swoogan/mongorest"
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

	NewMongoRest(db, "customers")
	NewMongoRest(db, "employees")

	log.Printf("About to listen on 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
