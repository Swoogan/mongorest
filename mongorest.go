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
	"github.com/Swoogan/rest.go"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
)

const (
	RW = iota
	RO
	WO
)

var formatting = "Valid JSON is required\n"

type Document map[string]interface{}

type created interface {
	Created(Document)
}

type updated interface {
	Updated(Document)
}

type Resource struct {
	DB mgo.Database
	Name string
	Mode int
	Handler interface{}
}

type MongoRest struct {
	col mgo.Collection
	log *log.Logger
	mode int
	handler interface{}
}

// Get all of the documents in the mongo collection 
func (mr *MongoRest) Index(w http.ResponseWriter, r *http.Request) {
	if mr.mode == WO {
		mr.log.Println("Attempt to read from write only resource")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var lookup Document
	if len(r.URL.RawQuery) > 0 {
		var err os.Error
		if lookup, err = parseQuery(r.URL.Query()); err != nil {
			mr.log.Println(err)
			rest.BadRequest(w, err.String())
			return
		}
	}

	var result []Document
	err := mr.col.Find(lookup).All(&result)
	if err != nil {
		mr.log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctype := contentType(r.Header.Get("accept"))
	switch ctype {
	case "application/json":
		enc := json.NewEncoder(w)
		w.Header().Set("content-type", ctype)
		enc.Encode(&result)
	case "text/html":
		w.Header().Set("content-type", ctype)
		//TODO: Implement templating here
		writeHtml(w, result)
	default:
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

// Find a document in the collection, identified by the ID
func (mr *MongoRest) Find(w http.ResponseWriter, idString string, r *http.Request) {
	if mr.mode == WO {
		mr.log.Println("Attempt to read from write only resource")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var result Document
	id := createIdLookup(idString)
	if err := mr.col.Find(id).One(&result); err != nil {
		mr.log.Println(err)
		rest.NotFound(w)
		return
	}

	ctype := contentType(r.Header.Get("accept"))
	switch ctype {
	case "application/json":
		enc := json.NewEncoder(w)
		w.Header().Set("content-type", ctype)
		enc.Encode(&result)
	case "text/html":
		w.Header().Set("content-type", ctype)
		//TODO: Implement templating here
		for key, value := range result {
			fmt.Fprintf(w, "%v: %v<br />", key, value)
		}
	default:
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

// Create and add a new document to the collection
func (mr *MongoRest) Create(w http.ResponseWriter, r *http.Request) {
	if mr.mode == RO {
		mr.log.Println("Attempt to write to read only resource")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctype := r.Header.Get("content-type")
	if ctype != "application/json" {
		mr.log.Println("Content type not implemented:", ctype)
		rest.NotImplemented(w)
		return
	}

	dec := json.NewDecoder(r.Body)
	var result Document
	if err := dec.Decode(&result); err != nil {
		mr.log.Println(err)
		//TODO: should this be a 406 or 415?
		rest.BadRequest(w, formatting)
		return
	}

	// Do insert
	if result["_id"] == nil {
		result["_id"] = bson.NewObjectId()
		if err := mr.col.Insert(result); err != nil {
			mr.log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if c, ok := mr.handler.(created); ok {
			c.Created(result)
		}
		output := fmt.Sprintf("%v%v", r.URL.String(), result["_id"])
		rest.Created(w, output)
		return
	}

	// Do upsert
	selector := bson.M{"_id": result["_id"]}
	if err := mr.col.Find(selector).One(&result); err != nil {
		if err2 := mr.col.Insert(result); err2 != nil {
			mr.log.Println(err2)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		output := fmt.Sprintf("%v%v", r.URL.String(), result["_id"])
		rest.Created(w, output)
		return
	}

	result["_id"] = nil, false
	change := bson.M{"$set": result}
	if err := mr.col.Update(selector, change); err != nil {
		mr.log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	/*
		var selector bson.M
		var change bson.M

		if result["_id"] == nil {
			change = result
		} else {
			selector = bson.M{"_id": result["_id"]}
			//result["_id"] = nil, false  // workaround for bug in mgo
			change = bson.M{"$set": result}
		}

		id, err := mr.col.Upsert(selector, change)

		switch {
		case err != nil:
			mr.log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		case id != nil:
			output := fmt.Sprintf("%v%v", r.URL.String(), id)
			rest.Created(w, output)
		default:
			w.WriteHeader(http.StatusOK)
		}
	*/
}

// Update a document identified by an ID with the data sent as request-body
func (mr *MongoRest) Update(w http.ResponseWriter, idString string, r *http.Request) {
	if mr.mode == RO {
		mr.log.Println("Attempt to write to read only resource")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctype := r.Header.Get("content-type")
	if ctype != "application/json" {
		mr.log.Println("Content type not implemented:", ctype)
		rest.NotImplemented(w)
		return
	}

	dec := json.NewDecoder(r.Body)
	var result Document

	if err := dec.Decode(&result); err != nil {
		mr.log.Println(err)
		rest.BadRequest(w, formatting)
		return
	}

	id := createIdLookup(idString)
	result["_id"] = id["_id"]

	// bug in mgo Upsert will fail if id is in 'result'
	newid, err := mr.col.Upsert(id, result)
	switch {
	case err != nil:
		mr.log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	case newid != nil:
		rest.Created(w, r.URL.String())
	default:
		rest.NoContent(w)
	}
}

// Delete a document identified by ID from the collection
func (mr *MongoRest) Delete(w http.ResponseWriter, idString string, r *http.Request) {
	if mr.mode == RO {
		mr.log.Println("Attempt to delete a read only resource")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := createIdLookup(idString)
	err := mr.col.Remove(id)
	if err == mgo.NotFound {
		// rest.NotFound(w)	// Deleting twice isn't supposed to be an error
		w.WriteHeader(http.StatusAccepted) // If it's delete, but we don't do anything, just accept it
	} else if err != nil {
		mr.log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		rest.NoContent(w)
	}
}

func Attach(res Resource, log *log.Logger) {
	mr := &MongoRest{res.DB.C(res.Name), log, res.Mode, res.Handler}
	rest.Resource(res.Name, mr)
}

