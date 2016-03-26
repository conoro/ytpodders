package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jmoiron/sqlx"
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
	var title string
	fmt.Println(strings.Join(args, " "))
	doc, err := goquery.NewDocument(args[0])
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		op, _ := s.Attr("property")
		con, _ := s.Attr("content")
		if op == "og:title" {
			fmt.Println(title)
			title = con
		}
	})
	db, err := sqlx.Connect("sqlite3", "ytpodders.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO subscriptions(suburl, subtitle, substatus) VALUES($1,$2,$3)", args[0], title, "enabled")
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
	RootCmd.AddCommand(AddCmd)
}
