package types

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type JsonStringSet map[string]bool

func StringSet(values ...string) JsonStringSet {
	var ans = make(JsonStringSet)
	for _, v := range values {
		ans[v] = true
	}
	return ans
}

func (s *JsonStringSet) MarshalJSON() ([]byte, error) {
	var keys = make([]string, 0, len(*s))
	for k := range *s {
		keys = append(keys, k)
	}
	return json.Marshal(keys)
}

func (s *JsonStringSet) UnmarshalJSON(bytes []byte) error {
	if *s == nil {
		*s = make(map[string]bool)
	}
	var keys []string
	err := json.Unmarshal(bytes, &keys)
	if err != nil {
		return err
	}
	for _, k := range keys {
		(*s)[k] = true
	}
	return nil
}

func (s *JsonStringSet) UnmarshalYAML(value *yaml.Node) error {
	var keys []string
	if err := value.Decode(&keys); err != nil {
		return err
	}
	if *s == nil {
		*s = make(map[string]bool, len(keys))
	}
	for _, k := range keys {
		(*s)[k] = true
	}
	return nil
}

func (s *JsonStringSet) MarshalYAML() (interface{}, error) {
	var keys = make([]string, 0, len(*s))
	for k := range *s {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *JsonStringSet) Has(key string) bool { return (*s)[key] }

func (s *JsonStringSet) Set(key string) {
	(*s)[key] = true
}

func (s *JsonStringSet) Del(key string) {
	delete(*s, key)
}

func (s *JsonStringSet) Dup() JsonStringSet {
	if s == nil {
		return nil
	}
	var cp = make(JsonStringSet, len(*s))
	for k, v := range *s {
		cp[k] = v
	}
	return cp
}
