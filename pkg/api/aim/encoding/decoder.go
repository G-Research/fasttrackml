package encoding

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
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
		return nil, eris.Wrap(err, "error reading data into the buffer")
	}
	data := make([]byte, binary.LittleEndian.Uint32(bufferLength))
	_, err = io.ReadFull(d, data)
	if err != nil {
		return nil, eris.Wrap(err, "error reading data into the buffer")
	}
	return data, nil
}

// DecoderProvider provides an interface to work with stream decoder.
type DecoderProvider interface {
	// Next implements iterator pattern and return data set by set.
	Next() (map[string]interface{}, error)
	// Decode decodes all the data stream at once.
	Decode() (map[string]interface{}, error)
}

// Decoder represents decoders for streaming data.
type Decoder struct {
	path   []string
	reader reader
	cursor string
}

// NewDecoder creates a new instance of Decoder.
func NewDecoder(data io.Reader) *Decoder {
	return &Decoder{
		reader: reader{bufio.NewReader(data)},
	}
}

// Next implements iterator pattern and return data set by set.
// nolint:gocyclo
// TODO:get back and fix `gocyclo` problem.
func (d *Decoder) Next() (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for {
		if len(d.path) == 0 {
			key, err := d.reader.readField()
			if err != nil {
				if err := eris.Unwrap(err); err == io.EOF {
					return result, err
				}
				return result, eris.Wrap(err, "error reading data line")
			}
			var index bool
			for _, p := range bytes.Split(key, []byte{0xFE}) {
				switch {
				case index:
					i := int64(binary.BigEndian.Uint64(p))
					d.path = append(d.path, strconv.FormatInt(i, 10))
					index = false
				case len(p) == 0:
					index = true
				default:
					d.path = append(d.path, string(p))
				}
			}
		}

		if len(d.path[0]) > 0 && d.cursor == "" {
			d.cursor = d.path[0]
		}

		// if `cursor` has been changed, then we have to release current object.
		if len(d.path) > 0 && d.cursor != d.path[0] {
			d.cursor = d.path[0]
			return result, nil
		}

		valuebuf, err := d.reader.readField()
		if err != nil {
			return nil, eris.Wrap(err, "error reading data line")
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

		result[strings.Join(d.path, ".")] = value
		d.path = []string{}
	}
}

// Decode decodes all the data stream at once. Duplicate keys are returned
// in the second return value.
func (d *Decoder) Decode() (map[string]interface{}, []string, error) {
	result := map[string]interface{}{}
	duplicates := []string{}
	for {
		data, err := d.Next()
		if len(data) > 0 {
			for key, value := range data {
				if _, ok := result[key]; ok {
					duplicates = append(duplicates, key)
				}
				result[key] = value
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return result, duplicates, nil
			}
			return nil, nil, eris.Wrap(err, "error decoding binary AIM stream")
		}
	}
}
