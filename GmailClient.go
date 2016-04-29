
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"encoding/base64"

"golang.org/x/net/context"
"golang.org/x/oauth2"
"golang.org/x/oauth2/google"
"google.golang.org/api/gmail/v1"
"github.com/calbucci/go-htmlparser"
)

type msgRetrievalError struct{
	errorMsg string
}

func (e *msgRetrievalError) Error() string{
	return e.errorMsg
}
// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
	"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gmail-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	link , _ := getResetPasswordLink()
	fmt.Println(link)
	//Code to fetch the link

}

func getResetPasswordLink() ( string, error) {
	var email = getContent("Reset Password")
	//fmt.Println(email)
	parser := htmlparser.NewParser(email)
	var links []string

	parser.Parse(nil, func(e *htmlparser.HtmlElement, isEmpty bool) {
		if e.TagName == "a" {
			link,_ := e.GetAttributeValue("href")
			if(link != "") {
				links = append(links, link)
			}
		}
	}, nil)

	if(len(links) > 0){
		return links[1], nil
	}else{
		return "", &msgRetrievalError{"no reset password link found"}
	}
}

func getVerifySignupLink() ( string, error) {
	var email = getContent("Verify Your Email")
	//fmt.Println(email)
	parser := htmlparser.NewParser(email)
	var links []string

	parser.Parse(nil, func(e *htmlparser.HtmlElement, isEmpty bool) {
		if e.TagName == "a" {
			link,_ := e.GetAttributeValue("href")
			if(link != "") {
				links = append(links, link)
			}
		}
	}, nil)

	if(len(links) > 0){
		return links[1], nil
	}else{
		return "", &msgRetrievalError{"no reset password link found"}
	}
}


func getContent(subject string) string {
	var foundMail string
	ctx := context.Background()
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}
	req, err := srv.Users.Messages.List("your email for which you created client client_secret").MaxResults(2).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}
	for _, li := range req.Messages {
		msg, err := srv.Users.Messages.Get("your email for which you created client client_secret", li.Id).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve message %v: %v", li.Id, err)
		}
		headers := msg.Payload.Headers
		for _, h := range headers {
			if (h.Name == "Subject") {
				switch subject {
				case "Reset Password":
					sDec, _ := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
					foundMail = string(sDec)
					break;
				case "Verify your Email":
					sDec, _ := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
					foundMail = string(sDec)
					break;
				default:

				}
			}
		}
	}
	return foundMail
}