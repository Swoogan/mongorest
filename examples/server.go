package main

import (
	"log"
	"http"
	"flag"
	"launchpad.net/mgo"
//	"bitbucket.org/Swoogan/mongorest"
	"mongorest"
)

func main() {
	var mongo *string = flag.String("m", "localhost", "Mongodb address")
	var dbname *string = flag.String("d", "test", "Mongodb database name")
	var address *string = flag.String("a", ":8080", "Address to listen on")
	flag.Parse()

	log.Printf("Connecting to mongodb at %v", *mongo)

	session, err := mgo.Mongo(*mongo)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	log.Printf("Opening database %v", *dbname)
	db := session.DB(*dbname)

	mongorest.New(db, "customers")
	mongorest.New(db, "employees")

	log.Printf("About to listen on %v", *address)
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
