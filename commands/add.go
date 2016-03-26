package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// AddCmd is the Action to run to add a YouTube Feed to your list
var AddCmd = &cobra.Command{
	Use:   "add url_of_youtube_user_or_channel",
	Short: "add a YouTube user URL or Channel URL to your ytpodders subscriptions",
	Long: `Pass the main Video URL of a User or Channel as the parameter e.g.

https://www.youtube.com/user/durianriders
or
https://www.youtube.com/channel/UCYdkEm-NjhS8TmLVt_qZy9g

From now on, when you run ytpodders, it will check that YouTuber
for any new entries.
`,
	Run: AddRun,
}

// AddRun is executed when user passes the command "add" to ytpodders
func AddRun(cmd *cobra.Command, args []string) {
	fmt.Println(strings.Join(args, " "))
}

func init() {
	RootCmd.AddCommand(AddCmd)
}
