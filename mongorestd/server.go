package main

import (
	"bitbucket.org/Swoogan/mongorest"
	"flag"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	session, err := mgo.Dial(*mongo)
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

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	for sig := range c {
		log.Printf("Received %v, shutting down...", sig)
		os.Exit(1)
	}
}

func createLogger(logfile string) *log.Logger {
	output := os.Stderr
	if logfile != "" {
		var err error
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

func setupResources(resources []string, db *mgo.Database, logger *log.Logger) {
	for _, resource := range resources {
		name, mode := parseResource(resource)
		switch mode {
		case "ro":
			res := mongorest.Resource{DB: db, Name: name, Mode: mongorest.RO}
			mongorest.Attach(res, logger)
			logger.Printf("Setting up resource '%v' in readonly mode", name)
		case "wo":
			res := mongorest.Resource{DB: db, Name: name, Mode: mongorest.WO}
			mongorest.Attach(res, logger)
			logger.Printf("Setting up resource '%v' in writeonly mode", name)
		default:
			res := mongorest.Resource{DB: db, Name: name}
			mongorest.Attach(res, logger)
			logger.Printf("Setting up resource '%v' in readwrite mode", name)
		}
	}
}
