package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/rotisserie/eris"
)

// supported list of reader types.
const (
	TypeNil = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
	TypeSlice
	TypeArray
	TypeObject
)

type reader struct {
	io.Reader
}

func (d *reader) readField() ([]byte, error) {
	bufferLength := make([]byte, 4)
	_, err := io.ReadFull(d, bufferLength)
	if err != nil {
		return nil, eris.Wrap(err, "Error reading Data into the buffer")
	}
	data := make([]byte, binary.LittleEndian.Uint32(bufferLength))
	_, err = io.ReadFull(d, data)
	if err != nil {
		return nil, eris.Wrap(err, "Error reading Data into the buffer")
	}
	return data, nil
}

type Decoder struct {
	data   reader
	cursor string
}

func NewDecoder(data io.Reader) *Decoder {
	return &Decoder{
		data: reader{bufio.NewReader(data)},
	}
}

// Decode decodes input stream of Data into map[string]interface{}.
// nolint:gocyclo
// TODO:get back and fix `gocyclo` problem.
func (d *Decoder) Decode() (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for {
		key, err := d.data.readField()
		if err != nil {
			if err := eris.Unwrap(err); err == io.EOF {
				return result, err
			}
			return result, eris.Wrap(err, "error reading Data line")
		}
		var index bool
		var path []string
		for _, p := range bytes.Split(key, []byte{0xFE}) {
			switch {
			case index:
				i := int64(binary.BigEndian.Uint64(p))
				path = append(path, strconv.FormatInt(i, 10))
				index = false
			case len(p) == 0:
				index = true
			default:
				path = append(path, string(p))
			}
		}

		if len(path) > 0 && d.cursor == "" {
			d.cursor = path[0]
		}

		// if `current` has been changed, then we have to release current object.
		if len(path) > 0 && d.cursor != path[0] {
			d.cursor = path[0]
			return result, nil
		}

		valuebuf, err := d.data.readField()
		if err != nil {
			return nil, eris.Wrap(err, "Error reading Data line")
		}

		var value any
		switch valuebuf[0] {
		case TypeNil:
			value = nil
		case TypeBool:
			value = valuebuf[1] != 0
		case TypeInt:
			switch len(valuebuf) - 1 {
			case 2:
				value = int16(binary.LittleEndian.Uint16(valuebuf[1:]))
			case 4:
				value = int32(binary.LittleEndian.Uint32(valuebuf[1:]))
			case 8:
				value = int64(binary.LittleEndian.Uint64(valuebuf[1:]))
			default:
				return nil, eris.Errorf("unsupported int length %d", len(valuebuf)-1)
			}
		case TypeFloat:
			switch len(valuebuf) - 1 {
			case 4:
				value = math.Float32frombits(binary.LittleEndian.Uint32(valuebuf[1:]))
			case 8:
				value = math.Float64frombits(binary.LittleEndian.Uint64(valuebuf[1:]))
			default:
				return nil, eris.Errorf("unsupported float length %d", len(valuebuf)-1)
			}
		case TypeString:
			value = string(valuebuf[1:])
		case TypeSlice:
			v := make([]float64, 0, (len(valuebuf)-1)/8)
			for i := 0; i < len(valuebuf)-1; i += 8 {
				v = append(v, math.Float64frombits(binary.LittleEndian.Uint64(valuebuf[i+1:])))
			}
			value = v
		case TypeArray:
			value = "<ARRAY>"
		case TypeObject:
			value = "<OBJECT>"
		default:
			return nil, eris.Errorf("unsupported type %x", valuebuf[0])
		}

		result[strings.Join(path, ".")] = value
	}
}
