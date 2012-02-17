package main

import (
	"os"
	"log"
	"http"
	"flag"
	"syscall"
	"os/signal"
	"launchpad.net/mgo"
	"bitbucket.org/Swoogan/mongorest"
)

var mongo *string = flag.String("m", "localhost", "Mongodb address")
var dbname *string = flag.String("d", "test", "Mongodb database name")
var address *string = flag.String("a", ":8080", "Address to listen on")
var logfile *string = flag.String("o", "", "File to log to")

func main() {
	flag.Parse()

	output := os.Stderr
	if *logfile != "" {
		var err os.Error
		output, err = os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
	logger := log.New(output, "", log.LstdFlags)

	if flag.NArg() == 0 {
		logger.Println("No resources specified, quiting...")
		return
	}

	logger.Printf("Connecting to mongodb at %v", *mongo)
	session, err := mgo.Mongo(*mongo)
	if err != nil {
		logger.Fatal(err)
	}
	defer session.Close()

	logger.Printf("Opening database %v", *dbname)
	db := session.DB(*dbname)

	for _, resource := range flag.Args() {
		mongorest.ReadWrite(db, resource, logger)
		logger.Println("Setting up resource:", resource)
	}

	logger.Printf("About to listen on %v", *address)
	go func() {
		err = http.ListenAndServe(*address, nil)
		if err != nil {
			logger.Fatal(err)
		}
	}()

	select {
	case sig := <-signal.Incoming:
		logger.Println("***Caught", sig)
		switch sig.(os.UnixSignal) {
		case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
			logger.Println("Shutting down...")
			return
		}
	}
}
