package fmp4io

import (
	"fmt"
	"strings"
)

type ParseError struct {
	Debug  string
	Offset int
	prev   *ParseError
	orig   error
}

func (a *ParseError) Error() string {
	s := []string{}
	for p := a; p != nil; p = p.prev {
		s = append(s, fmt.Sprintf("%s:%d", p.Debug, p.Offset))
		if p.prev == nil && p.orig != nil {
			s = append(s, p.orig.Error())
		}
	}
	return "mp4io: parse error: " + strings.Join(s, ",")
}

func parseErr(debug string, offset int, prev error) (err error) {
	_prev, _ := prev.(*ParseError)
	if _prev != nil {
		prev = nil
	}
	return &ParseError{
		Debug:  debug,
		Offset: offset,
		prev:   _prev,
		orig:   prev,
	}
}
