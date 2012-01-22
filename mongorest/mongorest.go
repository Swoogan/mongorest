package main

import (
	"fmt"
	"http"
	"log"
	"json"
	"os"
//	"io/ioutil"
//	"strconv"
//	"strings"
//	"xml"
	"github.com/nathankerr/rest.go"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
)

var formatting = "Valid JSON is required"

type MongoRest struct {
	col mgo.Collection
	json JsonDecoder
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

// Create and add a new document to the collection
func (mr *MongoRest) Create(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	result, err := mr.json.DecodeJson(dec);
	if err != nil {
		rest.BadRequest(w, formatting)
		return
	}

	result.newId()

	if err := mr.col.Insert(result); err != nil {
		rest.BadRequest(w, "later")
		return
        }

	rest.Created(w, fmt.Sprintf("%v%v", r.URL.String(), result.getId()))
}

// Update a document identified by an ID with the data sent as request-body
func (mr *MongoRest) Update(w http.ResponseWriter, idString string, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	result, err := mr.json.DecodeJson(dec);
	if err != nil {
		rest.BadRequest(w, formatting)
		return
	}

	// TODO: validate the Id first
	id := bson.ObjectIdHex(idString)

	err = mr.col.Update(bson.M{"_id": id}, result)
	if err == mgo.NotFound {
		rest.NotFound(w)
		return
	} else if err != nil {
		log.Println(err.String())
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

func NewMongoRest(db mgo.Database, resource string, json JsonDecoder) *MongoRest {
	mr :=  &MongoRest {
		col: db.C(resource),
		json: json,
	}

	rest.Resource(resource, mr)

	return mr
}

type Document interface {
	getId() string
	newId()
}

type JsonDecoder interface {
	DecodeJson(d *json.Decoder) (Document, os.Error)
}

type XmlDecoder interface {
	DecodeXml() (Document, os.Error)
}
