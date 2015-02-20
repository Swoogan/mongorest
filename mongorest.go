/*
A Go (golang) RESTful HTTP server library for exposing MongoDB document collections.
*/
package mongorest

import (
	"encoding/json"
	"fmt"
	"github.com/Swoogan/rest"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"strings"
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

type removed interface {
	Removed(bson.M)
}

type Resource struct {
	DB      *mgo.Database
	Name    string
	Mode    int
	Handler interface{}
}

type MongoRest struct {
	col     *mgo.Collection
	log     *log.Logger
	mode    int
	handler interface{}
}

// Get all of the documents in the mongo collection
func (mr *MongoRest) Index(w http.ResponseWriter, r *http.Request) {
	if mr.mode == WO {
		mr.log.Println("Attempt to read from write only resource")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var options queryOptions
	if len(r.URL.RawQuery) > 0 {
		var err error
		if options, err = parseQuery(r.URL.Query()); err != nil {
			mr.log.Println(err)
			rest.BadRequest(w, fmt.Sprint(err))
			return
		}
	}

	var result []Document
	query := mr.col.Find(options.criteria)
	iter := query.Select(options.selector)
	if err := iter.All(&result); err != nil {
		mr.log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mr.log.Println("Serving GET on", r.URL)

	mtype := mediaType(r.Header.Get("accept"))
	switch mtype {
	case "application/json":
		enc := json.NewEncoder(w)
		w.Header().Set("content-type", mtype)
		enc.Encode(&result)
	case "text/html":
		w.Header().Set("content-type", mtype)
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

	mr.log.Println("Serving GET on", r.URL)

	mtype := mediaType(r.Header.Get("accept"))
	switch mtype {
	case "application/json":
		enc := json.NewEncoder(w)
		w.Header().Set("content-type", mtype)
		enc.Encode(&result)
	case "text/html":
		w.Header().Set("content-type", mtype)
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
	if !strings.Contains(ctype, "application/json") {
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

	mr.log.Println("Serving POST on", r.URL)

	// Do insert
	if result["_id"] == nil {
		result["_id"] = bson.NewObjectId()
		mr.insert(w, r, result)
		return
	}

	// Do upsert...
	selector := bson.M{"_id": result["_id"]}
	if err := mr.col.Find(selector).One(&result); err != nil {
		mr.insert(w, r, result)
		return
	}

	// cont upsert
	result["_id"] = nil
	change := bson.M{"$set": result}
	if err := mr.col.Update(selector, change); err != nil {
		mr.log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if u, ok := mr.handler.(updated); ok {
		u.Updated(result)
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
	if !strings.Contains(ctype, "application/json") {
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

	mr.log.Println("Serving PUT on", r.URL)

	id := createIdLookup(idString)
	result["_id"] = id["_id"]

	// bug in mgo Upsert will fail if id is in 'result'
	newid, err := mr.col.Upsert(id, result)
	switch {
	case err != nil:
		mr.log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	case newid != nil:
		mr.insert(w, r, result)
		rest.Created(w, r.URL.String())
	default:
		if u, ok := mr.handler.(updated); ok {
			u.Updated(result)
		}
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

	mr.log.Println("Serving DELETE on", r.URL)

	id := createIdLookup(idString)
	switch err := mr.col.Remove(id); {
	case err == mgo.ErrNotFound:
		if d, ok := mr.handler.(removed); ok {
			d.Removed(id)
		}
		w.WriteHeader(http.StatusAccepted)
	case err != nil:
		mr.log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	default:
		rest.NoContent(w)
	}
}

func (mr *MongoRest) insert(w http.ResponseWriter, r *http.Request, doc Document) {
	if err := mr.col.Insert(doc); err != nil {
		mr.log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if c, ok := mr.handler.(created); ok {
		c.Created(doc)
	}
	output := fmt.Sprintf("%v%v", r.URL.String(), toString(doc["_id"]))
	mr.log.Println("New location: " + output)
	rest.Created(w, output)
}

func Attach(res Resource, log *log.Logger) {
	mr := &MongoRest{res.DB.C(res.Name), log, res.Mode, res.Handler}
	rest.Resource(res.Name, mr)
}
