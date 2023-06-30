package cmd

import (
	"fmt"
	"os"

	"github.com/rotisserie/eris"
	"github.com/spf13/cobra"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
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

		data, err := encoding.Decode(f)
		if err != nil {
			return eris.Wrap(err, "error decoding binary AIM stream")
		}

		fmt.Printf("decoded Aim stream data: %+v", data)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(DecodeCmd)
}
