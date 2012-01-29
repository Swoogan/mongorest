package main

import (
	"http"
	"log"
	"launchpad.net/mgo"
)

func main() {
	log.Printf("Connecting to mongodb")

	session, err := mgo.Mongo("172.16.1.63")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer session.Close()

	db := session.DB("uken")

	NewMongoRest(db, "finances")
	NewMongoRest(db, "games")

	log.Printf("About to listen on 4040")
	err = http.ListenAndServe(":4040", nil)
	if err != nil {
		log.Fatal(err)
	}
}
