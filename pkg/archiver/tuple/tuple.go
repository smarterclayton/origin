// Inspired in part by
// https://github.com/FoundationDB/fdb-go/blob/master/fdb/tuple/tuple.go
// and the SQLite key encoding schemes.  Copyright notice from the former
// included below
//
// Copyright (c) 2014 Red Hat, Inc
// Copyright (c) 2013 FoundationDB, LLC

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package tuple

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Convertible represents a type that can be converted into a key
type Convertible interface {
	Key() Key
}

// Key represents a key, a lexicographically-ordered sequence of
// bytes. Key implements the KeyConvertible interface.
type Key []byte

// The valid types for TupleElement are []byte (or Convertible), string,
// int64 (or int), and nil.
type TupleElement interface{}

// Tuple is a slice of objects that can be encoded as tuples. If
// any of the TupleElements are of unsupported types, a runtime panic will occur
// when the Tuple is packed.
//
// Given a Tuple T containing objects only of these types, then T will be
// identical to the Tuple returned by unpacking the byte slice obtained by
// packing T (modulo type normalization to []byte and int64).
type Tuple []TupleElement

var sizeLimits = []uint64{
	1<<(0*8) - 1,
	1<<(1*8) - 1,
	1<<(2*8) - 1,
	1<<(3*8) - 1,
	1<<(4*8) - 1,
	1<<(5*8) - 1,
	1<<(6*8) - 1,
	1<<(7*8) - 1,
	1<<(8*8) - 1,
}

func encodeBytes(buf *bytes.Buffer, code byte, b []byte) {
	buf.WriteByte(code)
	buf.Write(bytes.Replace(b, []byte{0x00}, []byte{0x00, 0xFF}, -1))
	buf.WriteByte(0x00)
}

func bisectLeft(u uint64) int {
	var n int
	for sizeLimits[n] < u {
		n += 1
	}
	return n
}

func encodeInt(buf *bytes.Buffer, i int64) {
	if i == 0 {
		buf.WriteByte(0x14)
		return
	}

	var n int
	var ibuf bytes.Buffer

	switch {
	case i > 0:
		n = bisectLeft(uint64(i))
		buf.WriteByte(byte(0x14 + n))
		binary.Write(&ibuf, binary.BigEndian, i)
	case i < 0:
		n = bisectLeft(uint64(-i))
		buf.WriteByte(byte(0x14 - n))
		binary.Write(&ibuf, binary.BigEndian, int64(sizeLimits[n])+i)
	}

	buf.Write(ibuf.Bytes()[8-n:])
}

// encodeUint64 encodes positive integers up to 64 bits.
func encodeUint64(buf *bytes.Buffer, i uint64) {
	if i == 0 {
		buf.WriteByte(0x14)
		return
	}

	var n int
	var ibuf bytes.Buffer

	n = bisectLeft(uint64(i))
	buf.WriteByte(byte(0x14 + n))
	binary.Write(&ibuf, binary.BigEndian, i)

	buf.Write(ibuf.Bytes()[8-n:])
}

// Pack returns a new byte slice encoding the provided tuple. Pack will panic if
// the tuple contains an element of any type other than []byte,
// Convertible, string, int64, int, uint64, or nil.
//
// Tuple satisfies the Convertible interface, so it is not necessary to
// call Pack when using a Tuple with an API function that requires a
// key.
func (t Tuple) Pack() []byte {
	buf := new(bytes.Buffer)

	for i, e := range t {
		switch e := e.(type) {
		case nil:
			buf.WriteByte(0x00)
		case int64:
			encodeInt(buf, e)
		case int:
			encodeInt(buf, int64(e))
		case uint64:
			encodeUint64(buf, e)
		case []byte:
			encodeBytes(buf, 0x01, e)
		case Convertible:
			encodeBytes(buf, 0x01, []byte(e.Key()))
		case string:
			encodeBytes(buf, 0x02, []byte(e))
		default:
			panic(fmt.Sprintf("unencodable element at index %d (%v, type %T)", i, t[i], t[i]))
		}
	}

	return buf.Bytes()
}

func findTerminator(b []byte) int {
	bp := b
	var length int

	for {
		idx := bytes.IndexByte(bp, 0x00)
		length += idx
		if idx+1 == len(bp) || bp[idx+1] != 0xFF {
			break
		}
		length += 2
		bp = bp[idx+2:]
	}

	return length
}

func decodeBytes(b []byte) ([]byte, int) {
	idx := findTerminator(b[1:])
	return bytes.Replace(b[1:idx+1], []byte{0x00, 0xFF}, []byte{0x00}, -1), idx + 2
}

func decodeString(b []byte) (string, int) {
	bp, idx := decodeBytes(b)
	return string(bp), idx
}

func decodeInt(b []byte) (int64, int) {
	if b[0] == 0x14 {
		return 0, 1
	}

	var neg bool

	n := int(b[0]) - 20
	if n < 0 {
		n = -n
		neg = true
	}

	bp := make([]byte, 8)
	copy(bp[8-n:], b[1:n+1])

	var ret int64

	binary.Read(bytes.NewBuffer(bp), binary.BigEndian, &ret)

	if neg {
		ret -= int64(sizeLimits[n])
	}

	return ret, n + 1
}

// Unpack returns the tuple encoded by the provided byte slice, or an error if
// the key does not correctly encode a tuple.
func Unpack(b []byte) (Tuple, error) {
	var t Tuple

	var i int

	for i < len(b) {
		var el interface{}
		var off int

		switch {
		case b[i] == 0x00:
			el = nil
			off = 1
		case b[i] == 0x01:
			el, off = decodeBytes(b[i:])
		case b[i] == 0x02:
			el, off = decodeString(b[i:])
		case 0x0c <= b[i] && b[i] <= 0x1c:
			el, off = decodeInt(b[i:])
		default:
			return nil, fmt.Errorf("unable to decode tuple element with unknown typecode %02x", b[i])
		}

		t = append(t, el)
		i += off
	}

	return t, nil
}

// Key returns the packed representation of a Tuple, and allows Tuple to
// satisfy the Convertible interface. Key will panic in the same
// circumstances as Pack.
func (t Tuple) Key() Key {
	return t.Pack()
}

func concat(a []byte, b ...byte) []byte {
	r := make([]byte, len(a)+len(b))
	copy(r, a)
	copy(r[len(a):], b)
	return r
}
