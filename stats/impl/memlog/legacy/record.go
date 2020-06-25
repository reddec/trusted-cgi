package legacy

import "time"

//go:generate msgp
//msgp:tuple Record

// Tracking record
type Record struct {
	UID    string    `json:"uid"`              // app UID
	Input  []byte    `json:"input,omitempty"`  // input data (could be empty if failed before read)
	Output []byte    `json:"output,omitempty"` // output data (could be empty if failed before run)
	Err    string    `json:"error,omitempty"`  // optional error
	Code   int       `json:"code"`             // response HTTP code
	Method string    `json:"method"`           // request HTTP method
	Remote string    `json:"remote"`           // request remote address (usually ip:port)
	Origin string    `json:"origin,omitempty"` // request origin header (could be empty)
	URI    string    `json:"uri"`              // raw request URI
	Token  string    `json:"token,omitempty"`  // request Authorization header (could be empty)
	Begin  time.Time `json:"begin"`            // started time
	End    time.Time `json:"end"`              // ended time
}
