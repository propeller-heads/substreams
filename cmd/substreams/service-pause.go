package main

import (
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/spf13/cobra"
	cli "github.com/streamingfast/cli"
	"github.com/streamingfast/cli/sflags"
	pbsinksvc "github.com/streamingfast/substreams/pb/sf/substreams/sink/service/v1"
	"github.com/streamingfast/substreams/pb/sf/substreams/sink/service/v1/pbsinksvcconnect"
	server "github.com/streamingfast/substreams/sink-server"
)

func init() {
	serviceCmd.AddCommand(pauseCmd)
}

var pauseCmd = &cobra.Command{
	Use:   "pause [deployment-id]",
	Short: "Pause a running service",
	Long: cli.Dedent(`
        Sends an "Pause" request to a server. By default, it will talk to a local "substreams service serve" instance.
        It will pause a substreams and returns information about the change of status.
        If deploymentID is not set or is incomplete, the CLI will try to guess (unless --strict is set).
		`),
	RunE:         pauseE,
	Args:         cobra.RangeArgs(0, 1),
	SilenceUsage: true,
}

func pauseE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	var id string
	if len(args) == 1 {
		id = args[0]
	}

	cli := pbsinksvcconnect.NewProviderClient(http.DefaultClient, sflags.MustGetString(cmd, "endpoint"))
	if len(id) < server.DeploymentIDLength {
		if sflags.MustGetBool(cmd, "strict") {
			return fmt.Errorf("invalid ID provided: %q and '--strict' is set", id)
		}
		matching, err := fuzzyMatchDeployment(ctx, id, cli, cmd, fuzzyMatchPreferredStatusOrder)
		if err != nil {
			return err
		}
		id = matching.Id
	}

	req := connect.NewRequest(&pbsinksvc.PauseRequest{
		DeploymentId: id,
	})
	if err := addHeaders(cmd, req); err != nil {
		return err
	}

	fmt.Printf("Pausing... please wait\n")
	resp, err := cli.Pause(ctx, req)
	if err != nil {
		return interceptConnectionError(err)
	}
	fmt.Printf("Response for deployment %q:\n  Previous Status: %v, New Status: %v\n", id, resp.Msg.PreviousStatus, resp.Msg.NewStatus)

	return nil
}
