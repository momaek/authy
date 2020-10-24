package cmd

import (
	"bufio"
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
}

var alfredCount *int

func init() {
	rootCmd.AddCommand(fuzzCmd)
	alfredCount = fuzzCmd.Flags().CountP("alfred", "a", "Specify Output Mode AlfredWorkflow")
}

func fuzzySearch(args []string) {
	var keyword string

	oc := bufio.NewScanner(os.Stdin)
	if len(args) == 0 {
		fmt.Print("Please input search keyword: ")
		if !oc.Scan() {
			log.Println("Please provide a keyword, e.g. google")
			return
		}

		args = []string{strings.TrimSpace(oc.Text())}
	}

	keyword = args[0]

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

	results := fuzzy.FindFrom(keyword, Tokens(tokens))
	if alfredCount != nil && *alfredCount > 0 {
		printAlfredWorkflow(results, tokens)
		return
	}

	prettyPrintResult(results, tokens)
}

func printAlfredWorkflow(results fuzzy.Matches, tokens []Token) {
	outputs := []AlfredOutput{}
	for _, v := range results {
		tk := tokens[v.Index]
		codes := totp.GetTotpCode(tk.Secret, tk.Digital)
		challenge := totp.GetChallenge()
		outputs = append(outputs, AlfredOutput{
			Title:    makeTitle(tk.Name, tk.OriginalName),
			Subtitle: makeSubTitle(challenge, codes[1]),
			Arg:      codes[1],
			Valid:    true,
		})
	}

	m := map[string][]AlfredOutput{"items": outputs}
	b, _ := json.Marshal(m)
	fmt.Println(string(b))
}

const (
	// Black black
	Black = "\033[1;30m%s\033[0m"
	// Red red
	Red = "\033[1;31m%s\033[0m"
	// Green green
	Green = "\033[1;32m%s\033[0m"
	// Yellow yellow
	Yellow = "\033[1;33m%s\033[0m"
	// Purple purple
	Purple = "\033[1;34m%s\033[0m"
	// Magenta magenta
	Magenta = "\033[1;35m%s\033[0m"
	// Teal teal
	Teal = "\033[1;36m%s\033[0m"
	// White white
	White = "\033[1;37m%s\033[0m"
	// DebugColor debug color
	DebugColor = "\033[0;36m%s\033[0m"
)

func prettyPrintResult(results fuzzy.Matches, tokens []Token) {
	fmt.Printf("\n")
	for _, r := range results {
		tk := tokens[r.Index]
		codes := totp.GetTotpCode(tk.Secret, tk.Digital)
		challenge := totp.GetChallenge()
		title := makeTitle(tk.Name, tk.OriginalName)
		fmt.Printf("- Title: "+Green+"\n", title)
		fmt.Printf("- Code: "+Teal+" Expires in "+Red+"(s)\n\n", codes[1], fmt.Sprint(calcRemainSec(challenge)))
	}
	return
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
