package mongorest

import (
	"os"
	"fmt"
	"log"
	"http"
	"strings"
	"strconv"
	"sre2.googlecode.com/hg/sre2"
	"launchpad.net/gobson/bson"
)

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

