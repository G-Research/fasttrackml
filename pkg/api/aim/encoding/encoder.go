package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"

	"github.com/rotisserie/eris"
)

type (
	arrayFlag  struct{}
	objectFlag struct{}
)

var pathSentinel = []byte{0xfe}

func EncodeTree(w io.Writer, tree map[string]any) error {
	return encodeTree(w, tree, []any{})
}

func encodeTree(w io.Writer, v any, p []any) error {
	if v == nil {
		return encodePathValue(w, v, p)
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return encodePathValue(w, v, p)
	case reflect.Slice, reflect.Array:
		// []byte
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return encodePathValue(w, v, p)
		}

		if err := encodePathValue(w, arrayFlag{}, p); err != nil {
			return err
		}
		for i := 0; i < rv.Len(); i++ {
			v := rv.Index(i).Interface()
			if err := encodeTree(w, v, append(p, i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		if rv.Len() == 0 {
			return encodePathValue(w, objectFlag{}, p)
		}

		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key().Interface()
			v := iter.Value().Interface()
			if err := encodeTree(w, v, append(p, k)); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported value %#v", v)
	}

	return nil
}

func encodePathValue(w io.Writer, v any, p []any) error {
	if err := encodePath(w, p); err != nil {
		return err
	}
	return encodeValue(w, v)
}

func encodePath(w io.Writer, p []any) error {
	buf := new(bytes.Buffer)
	for _, c := range p {
		switch c := c.(type) {
		case string:
			buf.WriteString(c)
			buf.Write(pathSentinel)
		case int:
			buf.Write(pathSentinel)
			if err := binary.Write(buf, binary.BigEndian, int64(c)); err != nil {
				return eris.Wrap(err, "error writing data into buffer")
			}
			buf.Write(pathSentinel)
		default:
			return fmt.Errorf("unsupported path component %#v", c)
		}
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(buf.Len())); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}

func encodeValue(w io.Writer, v any) error {
	var kind byte
	buf := new(bytes.Buffer)
	var bin bool
	switch v {
	case nil:
		kind = 0x00
	default:
		switch t := v.(type) {
		case bool:
			kind = 0x01
			bin = true
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			kind = 0x02
			bin = true
			switch t := v.(type) {
			case int:
				v = int64(t)
			case uint:
				v = uint64(t)
			}
		case float32, float64:
			kind = 0x03
			bin = true
		case string:
			kind = 0x04
			buf.WriteString(t)
		case []byte:
			kind = 0x05
			buf.Write(t)
		case arrayFlag:
			kind = 0x06
		case objectFlag:
			kind = 0x07
		default:
			return fmt.Errorf("unsupported value %#v", v)
		}
	}

	if bin {
		if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
			return eris.Wrap(err, "error writing data into buffer")
		}
	}

	if err := binary.Write(w, binary.LittleEndian, uint32(buf.Len()+1)); err != nil {
		return err
	}
	if _, err := w.Write([]byte{kind}); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}
