package api

import (
	"fmt"
	"net/url"
	"reflect"

	"golang.org/x/text/language"
)

type QueryArgs struct {
	IDs       []string          `url:"ids"`
	Languages []language.Tag    `url:"translatedLanguage"`
	Mangas    []string          `url:"manga"`
	Order     map[string]string `url:"order"`
	Limit     int               `url:"limit"`
	Offset    int               `url:"offset"`
}

func (a QueryArgs) Values() url.Values {
	result := make(url.Values)
	v := reflect.ValueOf(a)
	t := reflect.TypeOf(a)

	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Tag.Get("url")
		val := v.Field(i)
		if !val.IsZero() {
			if val.Kind() == reflect.Slice {
				for i := 0; i < val.Len(); i++ {
					result.Add(key+"[]", fmt.Sprint(val.Index(i)))
				}
			} else if val.Kind() == reflect.Map {
				for _, subkey := range val.MapKeys() {
					result.Add(key+"["+fmt.Sprint(subkey)+"]", fmt.Sprint(val.MapIndex(subkey)))
				}
			} else {
				result.Add(key, fmt.Sprint(val))
			}
		}
	}

	return result
}
