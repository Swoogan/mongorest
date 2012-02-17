package main

import (
	"os"
	"log"
	"http"
	"flag"
	"strings"
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

	logger := createLogger(*logfile)

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

	setupResources(flag.Args(), db, logger)

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

func createLogger(logfile string) *log.Logger {
	output := os.Stderr
	if logfile != "" {
		var err os.Error
		output, err = os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
	return log.New(output, "", log.LstdFlags)
}

func parseResource(resource string) (string, string) {
	res := strings.Split(resource, ",")
	name := res[0]
	mode := "rw"
	if len(res) > 1 && len(res[1]) == 2 {
		mode = res[1]
	}
	return name, mode
}

func setupResources(resources []string, db mgo.Database, logger *log.Logger) {
	for _, resource := range resources {
		name, mode := parseResource(resource)
		switch mode {
		case "ro":
			mongorest.ReadOnly(db, resource, logger)
			logger.Printf("Setting up resource '%v' in readonly mode", name)
		case "wo":
			mongorest.WriteOnly(db, resource, logger)
			logger.Printf("Setting up resource '%v' in writeonly mode", name)
		default:
			mongorest.ReadWrite(db, resource, logger)
			logger.Printf("Setting up resource '%v' in readwrite mode", name)
		}
	}
}
