package memlog

import (
	"github.com/reddec/trusted-cgi/stats"
	v0 "github.com/reddec/trusted-cgi/stats/impl/memlog/legacy"
	"github.com/reddec/trusted-cgi/types"
	"github.com/tinylib/msgp/msgp"
)

func isLegacyRecord(itemReader *msgp.Reader) (bool, error) {
	t, err := itemReader.NextType()
	if err != nil {
		return false, err
	}
	if t == msgp.ArrayType {
		return true, nil
	}
	return false, nil
}

func fromLegacy(reader *msgp.Reader) (*stats.Record, error) {
	var src v0.Record
	err := src.DecodeMsg(reader)
	if err != nil {
		return nil, err
	}
	return &stats.Record{
		UID: src.UID,
		Err: src.Err,
		Request: types.Request{
			Method:        src.Method,
			URL:           src.URI,
			Path:          "",
			RemoteAddress: src.Remote,
		},
		Begin: src.Begin,
		End:   src.End,
	}, nil
}
