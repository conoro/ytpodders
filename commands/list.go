package commands

import (
	"fmt"
	"os"

	"github.com/asdine/storm"
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
	db, err := storm.Open("ytpodders.boltdb", storm.AutoIncrement())
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	var ytSubscriptions []YTSubscription
	err = db.All(&ytSubscriptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, subscription := range ytSubscriptions {
		fmt.Println(subscription)
	}
}

func init() {
	RootCmd.AddCommand(ListCmd)
}
