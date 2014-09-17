package mongorest

import (
	"fmt"
	"regexp"
	"errors"
	"strings"
	"strconv"
	"net/http"
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
)

func toString(val interface{}) string {
	switch val.(type) {
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	case bson.ObjectId:
		o, _ := val.(bson.ObjectId)
		return o.Hex()
	}

	return "unknown"
}

func writeHtml(w http.ResponseWriter, items []Document) {
	for _, item := range items {
		for key, value := range item {
			fmt.Fprintf(w, "%v: %v<br />", key, value)
		}
		fmt.Fprint(w, "<br />")
	}
}

func createIdLookup(id string) bson.M {
	// TODO: add dates to this
	hex := regexp.MustCompile("^[0-9a-f]{24}$")
	number := regexp.MustCompile("^[0-9]+$")

	if hex.MatchString(id) {
		return bson.M{"_id": bson.ObjectIdHex(id)}
	} else if number.MatchString(id) {
		if i, err := strconv.Atoi(id); err == nil {
			return bson.M{"_id": i}
		}
	}

	return bson.M{"_id": id}
}

type queryOptions struct {
	criteria Document
	selector Document
	/*
	sort Document
	skip int
	limit int
	*/
}

func parseQuery(query map[string][]string) (queryOptions, error) {
	var err error
	var options queryOptions

	for key, values := range query {
		switch key {
			case "criteria":
				if len(values) > 1 {
					return options, errors.New("Can only have one criteria specified")
				}
				value := []byte(values[0])
				if err := json.Unmarshal(value, &options.criteria); err != nil {
					return options, err
				}
			case "selector":
				if len(values) > 1 {
					return options, errors.New("Can only have one selector specified")
				}
				value := []byte(values[0])
				if err := json.Unmarshal(value, &options.selector); err != nil {
					return options, err
				}
		}
	}

	return options, err
}

func parseAccept(accept string) []string {
	types := strings.Split(accept, ",")
	result := make([]string, 0)
	for _, atype := range types {
		typequal := strings.Split(atype, ";")
		clean := strings.TrimSpace(typequal[0])
		if !contains(result, clean) {
			result = append(result, clean)
		}
	}
	return result
}

func contains(haystack []string, needle string) bool {
	for _, val := range haystack {
		if val == needle {
			return true
		}
	}
	return false
}

func mediaType(accept string) string {
	if accept == "" {
		return "application/json"
	}

	var media string
	types := parseAccept(accept)
	switch {
	case contains(types, "application/json"):
		media = "application/json"
	case contains(types, "text/html"):
		media = "text/html"
	case contains(types, "application/*"):
		media = "application/json"
	case contains(types, "text/*"):
		media = "text/html"
	case contains(types, "*/*"):
		media = "application/json"
	}
	return media
}
