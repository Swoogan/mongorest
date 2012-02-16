/*
A Go (golang) RESTful HTTP server library for exposing MongoDB document collections.
*/
package mongorest

import (
	"os"
	"fmt"
	"log"
	"http"
	"json"
	"strings"
	"github.com/Swoogan/rest.go"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
)

var formatting = "Valid JSON is required"

type MongoRest struct {
	col mgo.Collection
}

// Get all of the documents in the mongo collection 
func (mr *MongoRest) Index(w http.ResponseWriter, r *http.Request) {
	var lookup map[string]interface{}
	if len(r.URL.RawQuery) > 0 {
		var err os.Error
		if lookup, err = parseQuery(r.URL.Query()); err != nil {
			rest.BadRequest(w, err.String())
			return
		}
	}

	var result []map[string]interface{}
	err := mr.col.Find(lookup).All(&result)
	if err != nil {
		log.Fatal(err)
	}

	switch accept := r.Header.Get("accept"); {
	case strings.Contains(accept, "application/json"):
		enc := json.NewEncoder(w)
		w.Header().Set("content-type", "application/json")
		enc.Encode(&result)
		//case strings.Contains(accept, "text/html"):
		//	w.Header().Set("content-type", "text/html")
		//	writeHtml(w, result)
	default:
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

// Find a document in the collection, identified by the ID
func (mr *MongoRest) Find(w http.ResponseWriter, idString string, r *http.Request) {
	var result map[string]interface{}
	id := createIdLookup(idString)
	if err := mr.col.Find(id).One(&result); err != nil {
		rest.NotFound(w)
		return
	}

	switch accept := r.Header.Get("accept"); {
	case strings.Contains(accept, "application/json"):
		enc := json.NewEncoder(w)
		w.Header().Set("content-type", "application/json")
		enc.Encode(&result)
		//case strings.Contains(accept, "text/html"):
		//	w.Header().Set("content-type", "text/html")
		//	for key, value := range result {
		//		fmt.Fprintf(w, "%v: %v<br />", key, value)
		//	}
	default:
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

// Create and add a new document to the collection
func (mr *MongoRest) Create(w http.ResponseWriter, r *http.Request) {
	ctype := r.Header.Get("content-type")
	if ctype != "application/json" {
		rest.NotImplemented(w)
		return
	}

	dec := json.NewDecoder(r.Body)
	var result map[string]interface{}
	if err := dec.Decode(&result); err != nil {
		rest.BadRequest(w, formatting)
		return
	}

	if result["_id"] == nil {
		result["_id"] = bson.NewObjectId()
	}

	if err := mr.col.Insert(result); err != nil {
		rest.BadRequest(w, "Could not insert document")
		return
	}

	output := fmt.Sprintf("%v%v", r.URL.String(), toString(result["_id"]))
	rest.Created(w, output)
}

// Update a document identified by an ID with the data sent as request-body
func (mr *MongoRest) Update(w http.ResponseWriter, idString string, r *http.Request) {
	ctype := r.Header.Get("content-type")
	if ctype != "application/json" {
		rest.NotImplemented(w)
		return
	}

	dec := json.NewDecoder(r.Body)
	var result map[string]interface{}

	if err := dec.Decode(&result); err != nil {
		rest.BadRequest(w, formatting)
		return
	}

	id := createIdLookup(idString)

	newid, err := mr.col.Upsert(id, result)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if newid != nil {
		if id, ok := newid.(bson.ObjectId); ok {
			rest.Created(w, id.Hex())
		} else {
			rest.NoContent(w)
		}
	} else {
		rest.NoContent(w)
	}
}

// Delete a snip identified by ID from the collection
func (mr *MongoRest) Delete(w http.ResponseWriter, idString string, r *http.Request) {
	id := createIdLookup(idString)
	err := mr.col.Remove(id)
	if err == mgo.NotFound {
		// rest.NotFound(w)	// Deleting twice isn't supposed to be an error
		w.WriteHeader(http.StatusAccepted) // If it's delete, but we don't do anything, just accept it
	} else if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		rest.NoContent(w)
	}
}

func New(db mgo.Database, resource string) *MongoRest {
	mr := &MongoRest{db.C(resource)}
	rest.Resource(resource, mr)
	return mr
}
