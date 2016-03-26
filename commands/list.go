package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// ListCmd is the Action to run to add a YouTube Feed to your ytpodders subscriptions
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "list all your ytpodder subscriptions",
	Long: `Get the name, URL and status of each of your subscriptions.
        `,
	Run: ListRun,
}

// ListRun is executed when user passes the command "list" to ytpodders
func ListRun(cmd *cobra.Command, args []string) {
	fmt.Println(strings.Join(args, " "))
}

func init() {
	RootCmd.AddCommand(ListCmd)
}
