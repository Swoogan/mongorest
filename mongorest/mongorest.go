package mongorest

import (
	"os"
	"fmt"
	"log"
	//	"url"
	"http"
	"json"
	"strings"
	"strconv"
	"github.com/Swoogan/rest.go"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
	"sre2.googlecode.com/hg/sre2"
)

var formatting = "Valid JSON is required"

func toString(val interface{}) string {
	switch v := val.(type) {
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	case bson.ObjectId:
		o, _ := val.(bson.ObjectId)
		return o.Hex()
	}

	return "unknown"
}

func writeHtml(w http.ResponseWriter, items []map[string]interface{}) {
	for _, item := range items {
		for key, value := range item {
			fmt.Fprintf(w, "%v: %v<br />", key, value)
		}
	}
}

func createIdLookup(id string) bson.M {
	// TODO: add dates to this
	hex := sre2.MustParse("^[0-9a-f]{24}$")
	number := sre2.MustParse("^[0-9]+$")

	if hex.Match(id) {
		return bson.M{"_id": bson.ObjectIdHex(id)}
	} else if number.Match(id) {
		if i, err := strconv.Atoi64(id); err == nil {
			return bson.M{"_id": i}
		}
	}

	return bson.M{"_id": id}
}

type MongoRest struct {
	col mgo.Collection
}

func parseQuery(query map[string][]string) (map[string]interface{}, os.Error) {
	var err os.Error
	result := make(map[string]interface{})

	for key, values := range query {
		if len(values) == 1 {
			result[key], err = convertType(values[0])
		} else if len(values) > 1 {
			log.Println("Arrays are handled with [a1,a2,...,an] syntax")
		}
	}

	return result, err
}

func convertType(value string) (interface{}, os.Error) {
	switch {
	case strings.Index(value, "s:") != -1:
		return value[2:], nil
	case strings.Index(value, "i:") != -1:
		return strconv.Atoi(value[2:])
	}

	// default to string
	return value, nil
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
	err := mr.col.Find(lookup).Limit(100).All(&result)
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

	//log.Println(r.Header.Get("content-type"))
	//log.Println(accept)
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
	dec := json.NewDecoder(r.Body)
	var result map[string]interface{}
	err := dec.Decode(&result)

	if err != nil {
		rest.BadRequest(w, formatting)
		return
	}

	id := createIdLookup(idString)

	err = mr.col.Update(id, result)
	if err == mgo.NotFound {
		rest.NotFound(w)
		return
	} else if err != nil {
		log.Println(err.String())
		// TODO: what to do if the doc doesn't update?
	}

	// Respond to indicate successful update
	rest.Updated(w, r.URL.String())
}

// Delete a snip identified by ID from the collection
func (mr *MongoRest) Delete(w http.ResponseWriter, idString string, r *http.Request) {
	id := createIdLookup(idString)
	err := mr.col.Remove(id)
	if err == mgo.NotFound {
		rest.NotFound(w)
		return
	} else if err != nil {
		log.Println(err.String())
		// TODO: what to do if the doc doesn't update?
	}

	rest.NoContent(w)
}

func NewMongoRest(db mgo.Database, resource string) *MongoRest {
	mr := &MongoRest{
		col: db.C(resource),
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
