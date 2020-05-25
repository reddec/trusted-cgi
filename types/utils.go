package types

import "encoding/json"

type JsonStringSet map[string]bool

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

func (s *JsonStringSet) Has(key string) bool { return (*s)[key] }

func (s *JsonStringSet) Set(key string) {
	(*s)[key] = true
}

func (s *JsonStringSet) Del(key string) {
	delete(*s, key)
}
