package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/brocaar/lora-geo-server/internal/test"
)

var testResolveTDOA = &cobra.Command{
	Use:     "test-resolve-tdoa",
	Short:   "Runs the resolve TDOA request from the given directory",
	Example: "lora-geo-server test-resolve-tdoa /path/to/request/logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("location to a file must be given as an argument")
		}

		/*

			b, err := ioutil.ReadFile(args[0])
			if err != nil {
				return errors.Wrap(err, "read file error")
			}

			var ts test.ResolveTDOATestSuite
			m := jsonpb.Unmarshaler{}
			if err = m.Unmarshal(bytes.NewReader(b), &ts); err != nil {
				return errors.Wrap(err, "unmarshal test file error")
			}

		*/

		return test.ResolveTDOA(args[0])
	},
}
