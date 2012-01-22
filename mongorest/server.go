package main

import (
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

type Game struct {
	Name string
	Id bson.ObjectId "_id,omitempty"
}

func (g *Game) getId() string {
	return g.Id.Hex()
}

func (g *Game) setId(id bson.ObjectId) {
	g.Id = id
}

type GameDecoder struct {}

func NewGameDecoder() *GameDecoder {
        return &GameDecoder{}
}

func (dr *GameDecoder) DecodeJson(d *json.Decoder) (Colly, os.Error) {
        var result Game
        err := d.Decode(&result)
        return &result, err
}

type Finance struct {
	Name string
	Id bson.ObjectId "_id,omitempty"
}

func (g *Finance) getId() string {
	return g.Id.Hex()
}

func (g *Finance) setId(id bson.ObjectId) {
	g.Id = id
}

type FinanceDecoder struct {}

func NewFinanceDecoder() *FinanceDecoder {
        return &FinanceDecoder{}
}

func (dr *FinanceDecoder) DecodeJson(d *json.Decoder) (Colly, os.Error) {
        var result Finance
        err := d.Decode(&result)
        return &result, err
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
	fd := NewFinanceDecoder()
	resource := NewMongoRest(col, fd)
	rest.Resource("finances", resource)

	col = session.DB("uken").C("games")
	gd := NewGameDecoder()
	resource = NewMongoRest(col, gd)
	rest.Resource("games", resource)

	log.Printf("About to listen on 4040")
	err = http.ListenAndServe(":4040", nil)
	if err != nil {
		log.Fatal(err)
	}
}
