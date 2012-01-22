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

func (g *Game) newId() {
	g.Id = bson.NewObjectId()
}

type GameJsonDecoder struct {}

func (dr GameJsonDecoder) DecodeJson(d *json.Decoder) (Document, os.Error) {
        var result Game
        err := d.Decode(&result)
        return &result, err
}

type Finance struct {
	Id bson.ObjectId "_id,omitempty"
	Balance float64
	Game string
	Income float64
	Wallet float64
}

func (g *Finance) getId() string {
	return g.Id.Hex()
}

func (g *Finance) newId() {
	g.Id = bson.NewObjectId()
}

type FinanceJsonDecoder struct {}

func (dr FinanceJsonDecoder) DecodeJson(d *json.Decoder) (Document, os.Error) {
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

	db := session.DB("uken")

	fd := FinanceJsonDecoder{}
	NewMongoRest(db, "finances", fd)

	gd := GameJsonDecoder{}
	NewMongoRest(db, "games", gd)

	log.Printf("About to listen on 4040")
	err = http.ListenAndServe(":4040", nil)
	if err != nil {
		log.Fatal(err)
	}
}
