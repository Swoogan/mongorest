package main

import (
	"fmt"
	"http"
	"log"
	"json"
//	"os"
//	"io/ioutil"
//	"strconv"
//	"strings"
//	"xml"
	"github.com/nathankerr/rest.go"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
)

var formatting = "formatting instructions go here"

type MongoRest struct {
	col mgo.Collection
}


// Get all of the documents in the mongo collection 
func (mr *MongoRest) Index(w http.ResponseWriter) {
	var result []interface{}
	if err := mr.col.Find(nil).Limit(100).All(&result); err != nil {
		panic(err)
	}

	enc := json.NewEncoder(w)
	w.Header().Set("content-type", "application/json")
	enc.Encode(&result)
}

// Find a snip from the collection, identified by the ID
func (mr *MongoRest) Find(w http.ResponseWriter, idString string) {
	var result interface{}
	id := bson.ObjectIdHex(idString)
	err := mr.col.Find(bson.M{"_id": id}).One(&result)
	if err != nil {
		rest.NotFound(w)
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Set("content-type", "application/json")
	enc.Encode(&result)
}

type Game struct {
	Name string
	_id bson.ObjectId
}

// Create and add a new document to the collection
func (mr *MongoRest) Create(w http.ResponseWriter, r *http.Request) {
	var result Game
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&result); err != nil {
		rest.BadRequest(w, err.String())
		return
	}

	if err := mr.col.Insert(result); err != nil {
		rest.BadRequest(w, "later")
		return
        }

	id := result._id
	rest.Created(w, fmt.Sprintf("%v%v", r.URL.String(), id))
}

// Update a snip identified by an ID with the data sent as request-body
func (mr *MongoRest) Update(w http.ResponseWriter, idString string, request *http.Request) {
	// Parse ID of type string to int
//	var id int
//	var err os.Error
//	if id, err = strconv.Atoi(idString); err != nil {
//		// The ID could not be converted from string to int
//		rest.NotFound(w)
//		return
//	}

	// Find the snip with the ID
//	var snip *Snip
//	var ok bool
//	if snip, ok = finances.WithId(id); !ok {
		// A snip with the passed ID could not be found in our collection
//		rest.NotFound(w)
//	}

	// Get the request-body for data to update the snipped to
//	var data []byte
//	if data, err = ioutil.ReadAll(request.Body); err != nil {
//		// The request body could not be read, thus it was a bad request
//		rest.BadRequest(w, formatting)
//		return
//	}

	// Set the finances body
//	snip.Body = string(data)
	// Respond to indicate successful update
	rest.Updated(w, request.URL.String())
}

// Delete a snip identified by ID from the collection
func (mr *MongoRest) Delete(w http.ResponseWriter, idString string) {
	id := bson.ObjectIdHex(idString)
	err := mr.col.Remove(bson.M{"_id": id})
	if err != nil {
		rest.NotFound(w)
	}

	rest.NoContent(w)
}

func NewMongoRest(col mgo.Collection) (*MongoRest) {
	return &MongoRest {col: col}
}

func main() {
	log.Printf("Connecting to mongodb")

	session, err := mgo.Mongo("172.16.1.63")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer session.Close()
	col := session.DB("uken").C("finances")

	resource := NewMongoRest(col)
	rest.Resource("finances", resource)

	col = session.DB("uken").C("games")
	resource = NewMongoRest(col)
	rest.Resource("games", resource)

	log.Printf("About to listen on 4040")
	err = http.ListenAndServe(":4040", nil)
	if err != nil {
		log.Fatal(err)
	}
}
