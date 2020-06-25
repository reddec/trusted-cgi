package stats

import (
	"github.com/reddec/trusted-cgi/types"
	"time"
)

//go:generate msgp

// Tracking record
type Record struct {
	UID     string        `json:"uid" msg:"uid,omitempty"`             // app UID
	Err     string        `json:"error,omitempty" msg:"err,omitempty"` // optional error
	Request types.Request `json:"request" msg:"req,omitempty"`         // incoming request
	Begin   time.Time     `json:"begin" msg:"beg,omitempty"`           // started time
	End     time.Time     `json:"end" msg:"end,omitempty"`             // ended time
}

// Recorder for apps requests
type Recorder interface {
	// Track single recorder
	Track(record Record)
}

// Reader from tracking systems. All returned records should be sorted from newest to oldest (by insertion moment)
type Reader interface {
	// Last records for specific app with limits
	LastByUID(uid string, limit int) ([]Record, error)
	// Last all records
	Last(limit int) ([]Record, error)
}

type Stats interface {
	Recorder
	Reader
}
