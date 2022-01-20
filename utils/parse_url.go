package utils

import (
	"fmt"
	"net/url"
	"strings"
)

type encoding int

const (
	encodePath encoding = 1 + iota
	encodePathSegment
	encodeHost
	encodeZone
	encodeUserPassword
	encodeQueryComponent
	encodeFragment
)

const upperhex = "0123456789ABCDEF"

func ParseURLToString(u *url.URL) string {
	var buf strings.Builder
	if u.Scheme != "" {
		buf.WriteString(u.Scheme)
		buf.WriteByte(':')
	}
	if u.Opaque != "" {
		buf.WriteString(u.Opaque)
	} else {
		if u.Scheme != "" || u.Host != "" || u.User != nil {
			if u.Host != "" || u.Path != "" || u.User != nil {
				buf.WriteString("//")
			}
			if ui := u.User; ui != nil {
				username := ui.Username()
				password, _ := ui.Password()
				buf.WriteString(fmt.Sprintf("%s:%s", username, password))
				buf.WriteByte('@')
			}
			if h := u.Host; h != "" {
				buf.WriteString(escape(h, encodeHost))
			}
		}
		path := u.EscapedPath()
		if path != "" && path[0] != '/' && u.Host != "" {
			buf.WriteByte('/')
		}
		if buf.Len() == 0 {
			if i := strings.IndexByte(path, ':'); i > -1 && strings.IndexByte(path[:i], '/') == -1 {
				buf.WriteString("./")
			}
		}
		buf.WriteString(path)
	}
	if u.ForceQuery || u.RawQuery != "" {
		buf.WriteByte('?')
		buf.WriteString(u.RawQuery)
	}
	if u.Fragment != "" {
		buf.WriteByte('#')
		buf.WriteString(u.EscapedFragment())
	}
	return buf.String()
}

func escape(s string, mode encoding) string {
	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c, mode) {
			if c == ' ' && mode == encodeQueryComponent {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		return s
	}

	var buf [64]byte
	var t []byte

	required := len(s) + 2*hexCount
	if required <= len(buf) {
		t = buf[:required]
	} else {
		t = make([]byte, required)
	}

	if hexCount == 0 {
		copy(t, s)
		for i := 0; i < len(s); i++ {
			if s[i] == ' ' {
				t[i] = '+'
			}
		}
		return string(t)
	}

	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ' && mode == encodeQueryComponent:
			t[j] = '+'
			j++
		case shouldEscape(c, mode):
			t[j] = '%'
			t[j+1] = upperhex[c>>4]
			t[j+2] = upperhex[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

func shouldEscape(c byte, mode encoding) bool {
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' {
		return false
	}

	if mode == encodeHost || mode == encodeZone {
		switch c {
		case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '[', ']', '<', '>', '"':
			return false
		}
	}

	switch c {
	case '-', '_', '.', '~':
		return false

	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@':
		switch mode {
		case encodePath:
			return c == '?'

		case encodePathSegment:
			return c == '/' || c == ';' || c == ',' || c == '?'

		case encodeUserPassword:
			return c == '@' || c == '/' || c == '?' || c == ':'

		case encodeQueryComponent:
			return true

		case encodeFragment:
			return false
		}
	}

	if mode == encodeFragment {
		switch c {
		case '!', '(', ')', '*':
			return false
		}
	}
	return true
}
