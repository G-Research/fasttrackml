package cmd

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var DecodeCmd = &cobra.Command{
	Use:    "decode",
	Short:  "Decodes a binary Aim stream",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		f := os.Stdin
		if len(args) > 0 {
			var err error
			f, err = os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()
		}

		d := decoder{bufio.NewReader(f)}

		for {
			key, err := d.readField()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
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
				return err
			}

			var value any
			switch valuebuf[0] {
			case 0:
				value = nil
			case 1:
				value = valuebuf[1] != 0
			case 2:
				switch len(valuebuf) - 1 {
				case 2:
					value = int16(binary.LittleEndian.Uint16(valuebuf[1:]))
				case 4:
					value = int32(binary.LittleEndian.Uint32(valuebuf[1:]))
				case 8:
					value = int64(binary.LittleEndian.Uint64(valuebuf[1:]))
				default:
					return fmt.Errorf("unsupported int length %d", len(valuebuf)-1)
				}
			case 3:
				switch len(valuebuf) - 1 {
				case 4:
					value = math.Float32frombits(binary.LittleEndian.Uint32(valuebuf[1:]))
				case 8:
					value = math.Float64frombits(binary.LittleEndian.Uint64(valuebuf[1:]))
				default:
					return fmt.Errorf("unsupported float length %d", len(valuebuf)-1)
				}
			case 4:
				value = string(valuebuf[1:])
			case 5:
				v := make([]float64, 0, (len(valuebuf)-1)/8)
				for i := 0; i < len(valuebuf)-1; i += 8 {
					v = append(v, math.Float64frombits(binary.LittleEndian.Uint64(valuebuf[i+1:])))
				}
				value = v
			case 6:
				value = "<ARRAY>"
			case 7:
				value = "<OBJECT>"
			default:
				return fmt.Errorf("unsupported type %x", valuebuf[0])
			}
			fmt.Printf("%s: %#v\n", strings.Join(path, "."), value)
		}
	},
}

type decoder struct {
	io.Reader
}

func (d *decoder) readField() ([]byte, error) {
	lenbuf := make([]byte, 4)
	_, err := io.ReadFull(d, lenbuf)
	if err != nil {
		return nil, err
	}
	len := binary.LittleEndian.Uint32(lenbuf)
	field := make([]byte, len)
	_, err = io.ReadFull(d, field)
	if err != nil {
		return nil, err
	}
	return field, nil
}

func init() {
	RootCmd.AddCommand(DecodeCmd)
}
