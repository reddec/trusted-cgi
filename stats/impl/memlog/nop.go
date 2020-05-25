package memlog

import (
	"github.com/reddec/trusted-cgi/stats"
)

func New(depth uint) *statLogger {
	return &statLogger{buffer: NewRingBuffer(depth)}
}

type statLogger struct {
	buffer *RingBuffer
}

func (s *statLogger) Track(record stats.Record) {
	s.buffer.Add(record)
}

func (s *statLogger) LastByUID(uid string, limit int) ([]stats.Record, error) {
	if limit < 0 {
		return []stats.Record{}, nil
	} else if n := s.buffer.Len(); limit > n {
		limit = n
	}
	var ans = make([]stats.Record, 0, limit)
	clone := s.buffer.Flatten()
	for i := len(clone) - 1; i >= 0 && len(ans) < limit; i-- {
		if clone[i].UID == uid {
			ans = append(ans, clone[i])
		}
	}
	return ans, nil
}

func (s *statLogger) Last(limit int) ([]stats.Record, error) {
	if limit < 0 {
		limit = 0
	}
	n := s.buffer.Len()
	if limit > n {
		limit = n
	}

	chunk := s.buffer.Flatten()[n-limit:]
	// order from newest to oldest
	for i, j := 0, len(chunk)-1; i < j; i, j = i+1, j-1 {
		chunk[i], chunk[j] = chunk[j], chunk[i]
	}

	return chunk, nil
}
