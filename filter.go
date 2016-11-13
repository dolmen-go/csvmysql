// Package csvmysql provides utilities for MySQL non-standard CSV format
package csvmysql

import (
	"errors"
	"io"
)

const SpecialNullString = "NULL\uFFFC"

var ErrSyntax = errors.New("unterminated escape")

func NewUnescapeReader(in io.Reader) *UnescapeReader {
	return &UnescapeReader{in: in}
}

type UnescapeReader struct {
	in     io.Reader
	buffer [4096]byte
	cur    []byte
}

// Replace:
//  `\N`  => NULL\uFFFC
//  `\\`  => \
//  `\"`  => ""
//  "\\n" => \n
func (r *UnescapeReader) Read(out []byte) (n int, err error) {
	if len(out) == 0 {
		return 0, nil
	}
	if len(r.cur) == 0 {
		n, err = r.in.Read(r.buffer[:cap(r.buffer)-len(SpecialNullString)])
		if n == 0 || (err != nil && err != io.EOF) {
			n = 0
			//log.Println(err)
			return
		}
		err = nil
		//log.Println(n)
		r.cur = r.buffer[:n]
	}

	j := len(r.cur)
	for i, c := range r.cur {
		if c == '\\' {
			j = i
			break
		}
	}
	if j >= 1 {
		n = copy(out, r.cur[:j])
		r.cur = r.cur[n:]
		//log.Printf("out: %s\n", out[:n])
		//log.Printf("cur: %s\n", r.cur)
		return
	}

	if len(r.cur) == 1 {
		// Must read one more byte
		n, err = r.in.Read(r.cur[1:2])
		if n != 1 {
			// Unterminated escape
			if err == io.EOF {
				return 0, ErrSyntax
			}
			return 0, err
		}
		r.cur = r.cur[:2]
		return
	}

	switch r.cur[1] {
	case 'N':
		// TODO Use r.buffer for storage
		tmp := make([]byte, len(SpecialNullString)+len(r.cur))
		copy(tmp[:len(SpecialNullString)], SpecialNullString)
		copy(tmp[len(SpecialNullString):], r.cur[2:])
		r.cur = tmp
		n = copy(out, r.cur)
		r.cur = r.cur[n:]
	case '"': // Replace `\"` with `""`
		r.cur[0] = '"'
		n = copy(out, r.cur[:2])
		r.cur = r.cur[n:]
	default: // Remove backslash
		out[0] = r.cur[1]
		r.cur = r.cur[2:]
		n = 1
	}
	//log.Printf("out: %s\n", out[:n])
	//log.Printf("cur: %s\n", r.cur)
	return
}
