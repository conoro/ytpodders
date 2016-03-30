package commands

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
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
	db, err := sqlx.Connect("sqlite3", "ytpodders.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM subscriptions WHERE ID=$1", args[0])
	tx.Commit()

	// query
	ytSubscriptions := []YTSubscription{}
	err = db.Select(&ytSubscriptions, "SELECT ID, suburl, subtitle, substatus FROM subscriptions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, subscription := range ytSubscriptions {
		fmt.Println(subscription)
	}
}

func init() {
	RootCmd.AddCommand(SubDeleteCmd)
}
