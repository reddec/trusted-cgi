package trace

import "io"

func NewSniffer(reader io.Reader, maxSniff int64) *Sniffer {
	return &Sniffer{
		max:    maxSniff,
		reader: reader,
	}
}

type Sniffer struct {
	max    int64
	total  int64
	data   []byte
	reader io.Reader
}

func (s *Sniffer) Data() []byte {
	return s.data
}

func (s *Sniffer) Total() int64 {
	return s.total
}

func (s *Sniffer) Report(trace *Trace, prefix string) {
	trace.Set(prefix+"_data", s.data)
	trace.Set(prefix+"_size", s.total)
}

func (s *Sniffer) Read(p []byte) (n int, err error) {
	if s.reader == nil {
		return 0, io.EOF
	}
	n, err = s.reader.Read(p)
	s.add(n, p)
	s.total += int64(n)
	return
}

func (s *Sniffer) add(n int, data []byte) {
	if s.max <= 0 {
		return
	}
	l := int64(n)
	if l > s.max {
		l = s.max
	}
	s.max -= l
	s.data = append(s.data, data[:l]...)
}
