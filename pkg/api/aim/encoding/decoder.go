package encoding

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

type DecoderResult struct {
	Data    map[string]interface{}
	Error   error
	Portion int
}

type DecoderProvider interface {
	// Decode represents syntactic sugar function which returns decoded stream data at once.
	// this function is just for back compatibility with integration tests.
	Decode() <-chan DecoderResult
	// DecodeByChunk decodes stream data by chunks.
	DecodeByChunk() (map[string]interface{}, error)
}

type Decoder struct {
	data io.Reader
}

func NewDecoder(data io.Reader) *Decoder {
	return &Decoder{data: data}
}

// DecodeByChunk decodes stream data by chunks.
func (d Decoder) DecodeByChunk() <-chan DecoderResult {
	portion, result := 0, make(chan DecoderResult)
	go func() {
		defer close(result)
		current, data, r := "", map[string]interface{}{}, reader{bufio.NewReader(d.data)}
		for {
			key, err := r.readField()
			if err != nil {
				if err := eris.Unwrap(err); err == io.EOF {
					result <- DecoderResult{
						Data:    data,
						Portion: portion,
					}
					return
				} else {
					result <- DecoderResult{
						Error: eris.Wrap(err, "error reading data line"),
					}
					return
				}
			}
			path, index := []string{}, false
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

			if len(path) > 0 && current == "" {
				current = path[0]
			}

			// if `current` has been changed, then we have to release current object.
			if len(path) > 0 && current != path[0] {
				current = path[0]
				result <- DecoderResult{
					Data:    data,
					Portion: portion,
				}
				data = map[string]interface{}{}
				portion++
			}

			valuebuf, err := r.readField()
			if err != nil {
				result <- DecoderResult{
					Error: eris.Wrap(err, "error reading data line"),
				}
				return
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
					result <- DecoderResult{
						Error: eris.Errorf("unsupported int length %d", len(valuebuf)-1),
					}
					return
				}
			case TypeFloat:
				switch len(valuebuf) - 1 {
				case 4:
					value = math.Float32frombits(binary.LittleEndian.Uint32(valuebuf[1:]))
				case 8:
					value = math.Float64frombits(binary.LittleEndian.Uint64(valuebuf[1:]))
				default:
					result <- DecoderResult{
						Error: eris.Errorf("unsupported int length %d", len(valuebuf)-1),
					}
					return
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
				result <- DecoderResult{
					Error: eris.Errorf("unsupported type %x", valuebuf[0]),
				}
				return
			}
			data[strings.Join(path, ".")] = value
		}
	}()
	return result
}

// Decode represents syntactic sugar function which returns decoded stream data at once.
// this function is just for back compatibility with integration tests.
func (d Decoder) Decode() (map[string]interface{}, error) {
	data := map[string]interface{}{}
	for result := range d.DecodeByChunk() {
		if result.Error != nil {
			return nil, eris.Wrap(result.Error, "error decoding binary AIM stream")
		} else {
			for key, value := range result.Data {
				data[key] = value
			}
		}
	}
	return data, nil
}

// Decode decodes input stream of Data into map[string]interface{}.
// nolint:gocyclo
// TODO:get back and fix `gocyclo` problem.
func Decode(data io.Reader) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	d := reader{bufio.NewReader(data)}
	for {
		key, err := d.readField()
		if err != nil {
			if err := eris.Unwrap(err); err == io.EOF {
				return result, nil
			}
			return result, eris.Wrap(err, "Error reading Data line")
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

		valuebuf, err := d.readField()
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
