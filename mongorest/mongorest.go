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

// Find a document in the collection, identified by the ID
func (mr *MongoRest) Find(w http.ResponseWriter, idString string) {
	var result interface{}
	// TODO: validate the Id first
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
	Id bson.ObjectId "_id,omitempty"
}

// Create and add a new document to the collection
func (mr *MongoRest) Create(w http.ResponseWriter, r *http.Request) {
	var result Game
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&result); err != nil {
		rest.BadRequest(w, err.String())
		return
	}

	result.Id = bson.NewObjectId()

	if err := mr.col.Insert(result); err != nil {
		rest.BadRequest(w, "later")
		return
        }

	rest.Created(w, fmt.Sprintf("%v%v", r.URL.String(), result.Id.Hex()))
}

// Update a document identified by an ID with the data sent as request-body
func (mr *MongoRest) Update(w http.ResponseWriter, idString string, r *http.Request) {
	var game Game
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&game); err != nil {
		rest.BadRequest(w, formatting)
		return
	}

	// TODO: validate the Id first
	id := bson.ObjectIdHex(idString)

	err := mr.col.Update(bson.M{"_id": id}, game)
	if err == mgo.NotFound {
		rest.NotFound(w)
		return
	} else if err != nil {
		// TODO: what to do if the doc doesn't insert?
	}

	// Respond to indicate successful update
	rest.Updated(w, r.URL.String())
}

// Delete a snip identified by ID from the collection
func (mr *MongoRest) Delete(w http.ResponseWriter, idString string) {
	// TODO: validate the Id first
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
