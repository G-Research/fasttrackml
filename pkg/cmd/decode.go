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
			//nolint:errcheck
			defer f.Close()
		}

		decoder := encoding.NewDecoder(f)
		fmt.Println("decoded Aim stream:")
		for result := range decoder.DecodeByChunk() {
			if result.Error != nil {
				return eris.Wrap(result.Error, "error decoding binary AIM stream")
			} else {
				for key, value := range result.Data {
					fmt.Printf("%s: %#v\n", key, value)
				}
				// time.Sleep(5 * time.Second)
			}
		}

		/*
			data, err := decoder.DecodeAll()
			if err != nil {
				return eris.Wrap(err, "error decoding binary AIM stream")
			}
			for key, value := range data {
				fmt.Printf("%s: %#v\n", key, value)
			}
		*/
		return nil
	},
}

func init() {
	RootCmd.AddCommand(DecodeCmd)
}
