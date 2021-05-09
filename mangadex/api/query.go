package api

import (
	"fmt"
	"net/url"
	"reflect"
)

type QueryArgs struct {
	IDs     []string `url:"ids[]"`
	MangaID string   `url:"manga"`
	Limit   int      `url:"limit"`
	Offset  int      `url:"offset"`
}

func (a QueryArgs) Values() url.Values {
	result := make(url.Values)
	v := reflect.ValueOf(a)
	t := reflect.TypeOf(a)

	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Tag.Get("url")
		switch f := v.Field(i).Interface().(type) {
		case []string:
			result[key] = f
		default:
			val := fmt.Sprint(f)
			if val != "" {
				result.Add(key, val)
			}
		}
	}

	return result
}
