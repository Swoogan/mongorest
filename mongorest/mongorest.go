package main

import (
	"fmt"
	"http"
	"log"
	"json"
	"os"
//	"io/ioutil"
//	"strconv"
	"strings"
//	"xml"
	"github.com/Swoogan/rest.go"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
)

var formatting = "Valid JSON is required"

type MongoRest struct {
	col mgo.Collection
	json JsonDecoder
}

func writeHtml(w http.ResponseWriter, items []map[string]interface{}) {
	for _, item := range items {
		for key, value := range item {
			fmt.Fprintf(w, "%v: %v<br />", key, value)
		}
	}
}

// Get all of the documents in the mongo collection 
func (mr *MongoRest) Index(w http.ResponseWriter, r *http.Request) {
	var result []map[string]interface{};
	err := mr.col.Find(nil).Limit(100).All(&result)
	if err != nil {
		panic(err)
	}

	switch accept := r.Header.Get("accept"); {
	case strings.Contains(accept, "application/json"):
		enc := json.NewEncoder(w)
		w.Header().Set("content-type", "application/json")
		enc.Encode(&result)
	case strings.Contains(accept, "text/html"):
		w.Header().Set("content-type", "text/html")
		writeHtml(w, result)
	default:
		w.WriteHeader(http.StatusNotAcceptable)
	}

//	log.Println(r.Header.Get("content-type"))
//	log.Println(accept)
}

// Find a document in the collection, identified by the ID
func (mr *MongoRest) Find(w http.ResponseWriter, idString string, r *http.Request) {
	var result map[string]interface{}
	// TODO: validate the Id first
	id := bson.ObjectIdHex(idString)
	err := mr.col.Find(bson.M{"_id": id}).One(&result)
	if err != nil {
		rest.NotFound(w)
		return
	}

	switch accept := r.Header.Get("accept"); {
	case strings.Contains(accept, "application/json"):
		enc := json.NewEncoder(w)
		w.Header().Set("content-type", "application/json")
		enc.Encode(&result)
	case strings.Contains(accept, "text/html"):
		w.Header().Set("content-type", "text/html")
		for key, value := range result {
			fmt.Fprintf(w, "%v: %v<br />", key, value)
		}
	default:
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

// Create and add a new document to the collection
func (mr *MongoRest) Create(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
        var result map[string]interface{}
        err := dec.Decode(&result)
	//result, err := mr.json.DecodeJson(dec);
	if err != nil {
		rest.BadRequest(w, formatting)
		return
	}

	result["_id"] = bson.NewObjectId()

	if err := mr.col.Insert(result); err != nil {
		rest.BadRequest(w, "Could not insert document")
		return
        }

	rest.Created(w, fmt.Sprintf("%v%v", r.URL.String(), result["_id"]))
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
func (mr *MongoRest) Delete(w http.ResponseWriter, idString string, r *http.Request) {
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
