package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/brocaar/chirpstack-geolocation-server/internal/test"
)

var testResolveMultiFrameTDOA = &cobra.Command{
	Use:     "test-resolve-multi-frame-tdoa",
	Short:   "Runs the resolve multi-frame TDOA request from the given directory",
	Example: "chirpstack-geolocation-server test-resolve-multi-frame-tdoa /path/to/request/logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("location to a file must be given as an argument")
		}

		return test.ResolveMultiFrameTDOA(args[0])
	},
}
