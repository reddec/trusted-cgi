package types

import (
	"encoding/json"
	"os"
	"time"
)

type Manifest struct {
	Name           string            `json:"name"`                      // information field
	Description    string            `json:"description"`               // information field
	Run            []string          `json:"run"`                       // command to run
	OutputHeaders  map[string]string `json:"output_headers"`            // output headers
	InputHeaders   map[string]string `json:"input_headers,omitempty"`   // headers to map from request to environment
	Query          map[string]string `json:"query,omitempty"`           // map query or form parameters to environment
	Environment    map[string]string `json:"environment,omitempty"`     // custom environment
	Method         string            `json:"method,omitempty"`          // restrict invoke only to the HTTP method
	MethodEnv      string            `json:"method_env,omitempty"`      // map method name to environment
	PathEnv        string            `json:"path_env,omitempty"`        // map requested path to environment
	TimeLimit      JsonDuration      `json:"time_limit,omitempty"`      // time limit to run (zero is infinity)
	MaximumPayload int64             `json:"maximum_payload,omitempty"` // limit incoming payload (zero is unlimited)
	AllowedIP      JsonStringSet     `json:"allowed_ip,omitempty"`      // limit incoming connections from list of IP
	AllowedOrigin  JsonStringSet     `json:"allowed_origin,omitempty"`  // limit incoming connections by origin header
	Public         bool              `json:"public"`                    // if public, tokens are ignores
	Tokens         map[string]string `json:"tokens,omitempty"`          // limit request by value in Authorization header (token => title)
}

func ExampleManifest() Manifest {
	return Manifest{
		Name: "example-app",
		Run:  []string{"echo", "<html><body><h1>Hello world</h1></body></html>"},
		OutputHeaders: map[string]string{
			"Content-Type": "text/html",
		},
	}
}

func (mf *Manifest) SaveAs(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(mf)
}

func (mf *Manifest) LoadFrom(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(mf)
}

type JsonDuration time.Duration

func (j *JsonDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(*j).String())
}

func (j *JsonDuration) UnmarshalJSON(bytes []byte) error {
	var str string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}
	v, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*j = JsonDuration(v)
	return nil
}
