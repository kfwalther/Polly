package auth

// List of imported packages
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"golang.org/x/oauth2"
)

type OAuthHandler struct {
	tokenFileName string
	oauthConfig   *oauth2.Config
}

// Create a new OAuth handler.
func NewOAuthHandler(tokenFile string, oauthConfig *oauth2.Config) *OAuthHandler {
	var handler OAuthHandler
	handler.tokenFileName = tokenFile
	handler.oauthConfig = oauthConfig
	return &handler
}

// Retrieve a token, saves the token, then returns the generated client.
func (h *OAuthHandler) GetHttpClient() *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := h.tokenFromFile()
	if err != nil {
		// Asynchronously get token from web. Token will be saved when redirect endpoint is hit.
		h.getTokenFromWeb()
		return nil
	}
	return h.oauthConfig.Client(context.Background(), tok)
}

// Method to handle the token response when OAuth flow returns Redirect URL.
func (h *OAuthHandler) HandleTokenResponse(w http.ResponseWriter, r *http.Request) {
	queryParts, _ := url.ParseQuery(r.URL.RawQuery)
	// Use the authorization code that is pushed to the redirect URL.
	authCode := queryParts["code"][0]
	log.Printf("code: %s\n", authCode)

	// Convert the authorization code into a token.
	token, err := h.oauthConfig.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	// Save the new token to the token file.
	h.saveToken(token)

	// Show a success page in the browser.
	msg := "<p><strong>Success!</strong></p>"
	msg = msg + "<p>You are authenticated and can now return to the CLI. You may close this tab.</p>"
	fmt.Fprintf(w, msg)
}

// Request a token from the web by opening the auth URL in a new browser window.
func (h *OAuthHandler) getTokenFromWeb() {
	authURL := h.oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	// Open the browser to authorize user with Google Sheets API via OAuth.
	exec.Command("rundll32", "url.dll,FileProtocolHandler", authURL).Start()
	log.Printf("Waiting for Google Sheets API auth token response before proceeding...")
}

// Retrieves a token from a local file.
func (h *OAuthHandler) tokenFromFile() (*oauth2.Token, error) {
	f, err := os.Open(h.tokenFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func (h *OAuthHandler) saveToken(token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", h.tokenFileName)
	f, err := os.OpenFile(h.tokenFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
