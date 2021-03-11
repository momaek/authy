package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	//"time"

	"github.com/alexzorin/authy"
	"github.com/momaek/authy/totp"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// fuzzCmd represents the fuzz command
var fuzzCmd = &cobra.Command{
	Use:   "fuzz",
	Short: "Fuzzy search your otp tokens(case-insensitive)",
	Long: `Fuzzy search your otp tokens(case-insensitive)

First time(or after clean cache) , need your authy main password`,
	Run: func(cmd *cobra.Command, args []string) {
		fuzzySearch(args)
	},
}

// Token save in cache
type Token struct {
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	Digital      int    `json:"digital"`
	Secret       string `json:"secret"`
	Period       int    `json:"period"`
}

// AlfredOutput alfred workflow output
type AlfredOutput struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Arg      string `json:"arg"`
	Icon     struct {
		Type string `json:"type"`
		Path string `json:"path"`
	} `json:"icon"`
	Valid bool `json:"valid"`
	Text  struct {
		Copy string `json:"copy"`
	} `json:"text"`

	Code      string `json:"-"`
	ExpireSec int    `json:"-"`
}

var alfredCount *int

func init() {
	rootCmd.AddCommand(fuzzCmd)
	alfredCount = fuzzCmd.Flags().CountP("alfred", "a", "Specify Output Mode AlfredWorkflow")
}

func fuzzySearch(args []string) {
	var (
		keyword string
	)

	if len(args) != 0 {
		keyword = args[0]
	}

	showResultByKeyword(keyword)
}

func getResult(keyword string) (foundTokens []Token, err error) {
	devInfo, err := LoadExistingDeviceInfo()
	if err != nil {
		if os.IsNotExist(err) {
			devInfo, err = newRegistrationDevice()
			if err != nil {
				return
			}
		} else {
			log.Println("load device info failed", err)
			return
		}
	}

	tokens, err := loadCachedTokens()
	if err != nil {
		tokens, err = getTokensFromAuthyServer(&devInfo)
		if err != nil {
			log.Fatal("get tokens failed", err)
		}
	}

	// default show all totp codes
	foundTokens = tokens
	if len(keyword) == 0 {
		results := fuzzy.FindFrom(keyword, Tokens(tokens))
		for _, v := range results {
			foundTokens = append(foundTokens, tokens[v.Index])
		}
	}

	return
}

func showResultByKeyword(keyword string) {
	foundTokens, err := getResult(keyword)
	if err != nil {
		return
	}

	outputs := []AlfredOutput{}
	for _, tk := range foundTokens {
		if len(tk.Secret) == 0 {
			break
		}

		codes := totp.GetTotpCode(tk.Secret, tk.Digital)
		challenge := totp.GetChallenge()
		outputs = append(outputs, AlfredOutput{
			Title:    makeTitle(tk.Name, tk.OriginalName),
			Subtitle: makeSubTitle(challenge, codes[1]),
			Arg:      codes[1],
			Valid:    true,

			Code:      codes[1],
			ExpireSec: calcRemainSec(challenge),
		})
	}

	if len(outputs) == 0 {
		outputs = append(outputs, AlfredOutput{
			Title:    "Please Refresh Autht Cache",
			Subtitle: "Run`authy delpwd && authy refresh`in Commandline. Press Enter copy",
			Arg:      "authy delpwd && authy refresh",
			Valid:    true,
		})
	}

	if alfredCount != nil && *alfredCount > 0 {
		printAlfredWorkflow(outputs)
		return
	}

	prettyPrintResult(outputs)
}

func printAlfredWorkflow(outputs []AlfredOutput) {
	m := map[string][]AlfredOutput{"items": outputs}
	b, _ := json.Marshal(m)
	fmt.Println(string(b))
}

func calcRemainSec(challenge int64) int {
	return 30 - int(time.Now().Unix()-challenge*30)
}

func makeSubTitle(challenge int64, code string) string {
	return fmt.Sprintf("Code: %s [Press Enter copy to clipboard], Expires in %d second(s)", code, calcRemainSec(challenge))
}

func makeTitle(name, originName string) string {
	if len(name) > len(originName) {
		return name
	}

	return originName
}

// Tokens for
type Tokens []Token

func (ts Tokens) String(i int) string {
	if len(ts[i].Name) > len(ts[i].OriginalName) {
		return ts[i].Name
	}

	return ts[i].OriginalName
}

// Len implement fuzzy.Source
func (ts Tokens) Len() int { return len(ts) }

const cacheFileName = ".authycache.json"

func loadCachedTokens() (tks []Token, err error) {
	fpath, err := ConfigPath(cacheFileName)
	if err != nil {
		return
	}

	f, err := os.Open(fpath)
	if err != nil {
		return
	}

	defer f.Close()
	err = json.NewDecoder(f).Decode(&tks)
	return
}

func saveTokens(tks []Token) (err error) {
	regrPath, err := ConfigPath(cacheFileName)
	if err != nil {
		return
	}

	f, err := os.OpenFile(regrPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return
	}

	defer f.Close()
	err = json.NewEncoder(f).Encode(&tks)
	return
}

func getTokensFromAuthyServer(devInfo *DeviceRegistration) (tks []Token, err error) {
	client, err := authy.NewClient()
	if err != nil {
		log.Fatalf("Create authy API client failed %+v", err)
	}

	apps, err := client.QueryAuthenticatorApps(nil, devInfo.UserID, devInfo.DeviceID, devInfo.Seed)
	if err != nil {
		log.Fatalf("Fetch authenticator apps failed %+v", err)
	}

	if !apps.Success {
		log.Fatalf("Fetch authenticator apps failed %+v", apps)
	}

	tokens, err := client.QueryAuthenticatorTokens(nil, devInfo.UserID, devInfo.DeviceID, devInfo.Seed)
	if err != nil {
		log.Fatalf("Fetch authenticator tokens failed %+v", err)
	}

	if !tokens.Success {
		log.Fatalf("Fetch authenticator tokens failed %+v", tokens)
	}

	if len(devInfo.MainPassword) == 0 {
		fmt.Print("\nPlease input Authy main password: ")
		pp, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Get password failed %+v", err)
		}

		devInfo.MainPassword = strings.TrimSpace(string(pp))
		SaveDeviceInfo(*devInfo)
	}

	tks = []Token{}
	for _, v := range tokens.AuthenticatorTokens {
		secret, err := v.Decrypt(devInfo.MainPassword)
		if err != nil {
			log.Fatalf("Decrypt token failed %+v", err)
		}

		tks = append(tks, Token{
			Name:         v.Name,
			OriginalName: v.OriginalName,
			Digital:      v.Digits,
			Secret:       secret,
		})
	}

	for _, v := range apps.AuthenticatorApps {
		secret, err := v.Token()
		if err != nil {
			log.Fatal("Get secret from app failed", err)
		}

		tks = append(tks, Token{
			Name:    v.Name,
			Digital: v.Digits,
			Secret:  secret,
			Period:  10,
		})
	}

	saveTokens(tks)
	return
}
