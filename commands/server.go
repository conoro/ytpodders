package commands

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/stacktic/dropbox"
	"golang.org/x/oauth2"
)

var db *dropbox.Dropbox
var dropboxLink *dropbox.Link

// UserToken is the Dropbox token
var UserToken string
var err error

// ServerConfiguration from conf.json
type ServerConfiguration struct {
	ClientID         string `json:"clientid"`
	ClientSecret     string `json:"clientsecret"`
	ServerURL        string `json:"serverurl"`
	ServerPort       string `json:"serverport"`
	OauthStateString string `json:"oauthstatestring"`
}

var (
	// Set callback to http://127.0.0.1:7000/dropbox_oauth_cb
	// Set ClientId and ClientSecret to
	oauthConf = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "",
		Scopes:       []string{""},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/1/oauth2/authorize",
			TokenURL: "https://api.dropbox.com/1/oauth2/token",
		},
	}
	// random string for oauth2 API calls to protect against CSRF
	oauthStateString = ""
)

// ServerCmd is the Action to run to run a Server to Authorise the App to use Dropbox
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "run a server to enable users to authorise YTPodders to use Dropbox",
	Long: `Facilitates the Dropbox OAuth flow so you can retrieve a token as a user and save that for future use by YTPodders
`,
	Run: ServerRun,
}

// ServerRun is executed when user passes the command "server" to ytpodders
func ServerRun(cmd *cobra.Command, args []string) {

	// Read configuration from conf.json
	conffile, _ := os.Open("server_conf.json")
	decoder := json.NewDecoder(conffile)
	config := ServerConfiguration{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("config error:", err)
		log.Fatal(err)
	}

	oauthConf.ClientID = config.ClientID
	oauthConf.ClientSecret = config.ClientSecret
	oauthConf.RedirectURL = config.ServerURL + "/dropbox_oauth_cb"
	oauthStateString = config.OauthStateString

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	//http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleDropboxLogin)
	http.HandleFunc("/dropbox_oauth_cb", handleDropboxCallback)
	http.HandleFunc("/success", handleSuccess)
	fmt.Print("Started running on http://127.0.0.1\n")
	fmt.Println(http.ListenAndServe(":"+config.ServerPort, nil))

}

// /login
func handleDropboxLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// /dropbox_oauth_cb. Called by Dropbox after authorization is granted
func handleDropboxCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	fmt.Printf("Save this token in client_conf.json: %s\n", token.AccessToken)
	UserToken = token.AccessToken
	http.Redirect(w, r, "/success", http.StatusTemporaryRedirect)
}

// /dropbox_oauth_cb. Called by Dropbox after authorization is granted
func handleSuccess(w http.ResponseWriter, r *http.Request) {
	type Page struct {
		UserToken string
	}

	p := Page{
		UserToken: UserToken,
	}

	tmpl, err := template.ParseFiles("templates/success.html.tmpl") // Parse template file.
	if err != nil {
		fmt.Printf("Error rendering template with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	tmpl.Execute(w, p)
}

func init() {
	RootCmd.AddCommand(ServerCmd)
}
