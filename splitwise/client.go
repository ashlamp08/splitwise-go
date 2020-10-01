package splitwise

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+-0123456789")

type Client struct {
	*RestClient
	conf   *oauth2.Config
	state  string
	logger log.Logger
	token  *oauth2.Token
}

func New(consumerKey, secret, redirectUrl string, httpClient http.Client) *Client {
	return &Client{
		RestClient: &RestClient{
			HttpClient: &httpClient,
		},
		conf: &oauth2.Config{
			RedirectURL:  redirectUrl,
			ClientID:     consumerKey,
			ClientSecret: secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  AuthorizeUrl,
				TokenURL: TokenUrl,
			},
		},
		logger: log.Logger{},
	}
}

func (c *Client) handleLogin(w http.ResponseWriter, r *http.Request) {

}

func (c *Client) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != "randomState" {
		fmt.Println("Invalid state!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, err := c.conf.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		fmt.Println("failed to get token due to : ", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://secure.splitwise.com/api/v3.0/get_current_user?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Println("failed to GET current user from Splitwise due to : ", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read all from response body due to : ", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Println("current_user : ", string(bytes))
}

func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	url := c.accessTokenToUrl(GetCurrentUserUrl)
	fmt.Println("GetCurrentUser url : ", url)
	resp, err := c.Get(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	var user User
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(bts, &user); err != nil {
		return nil, err
	}
	return &user, err
}

func (c *Client) accessTokenToUrl(url string) string {
	return url + "?access_token=" + c.token.AccessToken
}

func generateRandomState() string {
	// [65, 122]
	rand.Seed(time.Now().UnixNano())
	min, max := 65, 122
	s := ""
	for i := 0; i < 32; i++ {
		s += string(rune(rand.Intn(57) + 48))
	}
	fmt.Println("generated state s : ", s)
	fmt.Println(rand.Intn(max - min + 1) + min)
	return s
}
