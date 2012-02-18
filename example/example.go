package main

import (
	"os"
	"log"
	"http"
	"flag"
	"launchpad.net/mgo"
	//	"bitbucket.org/Swoogan/mongorest"
	"mongorest"
)

var mongo *string = flag.String("m", "localhost", "Mongodb address")
var dbname *string = flag.String("d", "test", "Mongodb database name")
var address *string = flag.String("a", ":8080", "Address to listen on")

func main() {
	flag.Parse()

	logger := log.New(os.Stderr, "", log.LstdFlags)
	logger.Printf("Connecting to mongodb at %v", *mongo)

	session, err := mgo.Mongo(*mongo)
	if err != nil {
		logger.Fatal(err)
	}
	defer session.Close()

	logger.Printf("Opening database %v", *dbname)
	db := session.DB(*dbname)

	cust := mongorest.Resource{DB: db, Name: "customers"}
	mongorest.Attach(cust, logger)
	emp := mongorest.Resource{DB: db, Name: "employees"}
	mongorest.Attach(emp, logger)

	logger.Printf("About to listen on %v", *address)
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		logger.Fatal(err)
	}
}
