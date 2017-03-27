package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/asdine/storm"
	"github.com/spf13/cobra"
)

// SubDeleteCmd is the Action to run to add a YouTube Feed to your list
var SubDeleteCmd = &cobra.Command{
	Use:   "delete ID_of_subscription",
	Short: "delete a ytpodders subscription",
	Long: ` use ytpodders list to get all the ids and ytpodders delete to delete one
`,
	Run: SubDeleteRun,
}

// SubDeleteRun is executed when user passes the command "add" to ytpodders
func SubDeleteRun(cmd *cobra.Command, args []string) {
	db, err := storm.Open("ytpodders.boltdb", storm.AutoIncrement())
	if err != nil {
		fmt.Fprintf(os.Stderr, "1 error: %v\n", err)
	}
	defer db.Close()

	var ytSubscription YTSubscription
	delID, _ := strconv.Atoi(args[0])
	err = db.One("ID", delID, &ytSubscription)
	if err != nil && err.Error() != "bucket YTSubscription not found" && err.Error() != "bucket YTSubscriptionEntry not found" {
		fmt.Fprintf(os.Stderr, "2 error: %v\n", err)
		os.Exit(1)
	}

	err = db.Remove(&ytSubscription)
	if err != nil {
		fmt.Fprintf(os.Stderr, "3 error: %v\n", err)
		os.Exit(1)
	}

	// query
	var ytSubscriptions []YTSubscription
	err = db.All(&ytSubscriptions)
	if err != nil && err.Error() != "bucket YTSubscription not found" && err.Error() != "bucket YTSubscriptionEntry not found" {
		fmt.Fprintf(os.Stderr, "4 error: %v\n", err)
		os.Exit(1)
	}

	for _, subscription := range ytSubscriptions {
		fmt.Println(subscription)
	}
}

func init() {
	RootCmd.AddCommand(SubDeleteCmd)
}
