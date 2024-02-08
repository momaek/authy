package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alexzorin/authy"
	"golang.org/x/crypto/ssh/terminal"
)

// Tokens for sort
type Tokens []*Token

func (t Tokens) Less(i, j int) bool {
	return t[i].Weight > t[j].Weight
}

func (t Tokens) Len() int {
	return len(t)
}

func (t Tokens) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// String implement fuzz search
func (t Tokens) String(i int) string {
	return t[i].Name + t[i].OriginalName
}

// Token ..
type Token struct {
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	Digital      int    `json:"digital"`
	Secret       string `json:"secret"`
	Period       int    `json:"period"`
	Weight       int    `json:"weight"`
}

// Title show string
func (t Token) Title() string {
	if t.Name != t.OriginalName || len(t.OriginalName) == 0 {
		return t.Name
	}

	return t.OriginalName
}

func (t *Token) updateWeight() {
	t.Weight++
}

// LoadTokenFromCache load token from local cache
func (d *Device) LoadTokenFromCache() (err error) {
	defer func() {
		if err != nil {
			d.LoadTokenFromAuthyServer()
			err = nil
		}
	}()

	fpath, err := d.ConfigPath(d.conf.CacheFileName)
	if err != nil {
		return
	}

	f, err := os.Open(fpath)
	if err != nil {
		return
	}

	defer f.Close()
	err = json.NewDecoder(f).Decode(&d.tokens)
	if err != nil {
		return
	}

	d.tokenMap = tokensToMap(d.tokens)

	return
}

// LoadTokenFromAuthyServer load token from authy server, make sure that you've enabled Authenticator Backups And Multi-Device Sync
func (d *Device) LoadTokenFromAuthyServer() {
	client, err := authy.NewClient()
	if err != nil {
		log.Fatalf("Create authy API client failed %+v", err)
	}

	apps, err := client.QueryAuthenticatorApps(nil, d.registration.UserID, d.registration.DeviceID, d.registration.Seed)
	if err != nil {
		log.Fatalf("Fetch authenticator apps failed %+v", err)
	}

	if !apps.Success {
		log.Fatalf("Fetch authenticator apps failed %+v", apps)
	}

	tokens, err := client.QueryAuthenticatorTokens(nil, d.registration.UserID, d.registration.DeviceID, d.registration.Seed)
	if err != nil {
		log.Fatalf("Fetch authenticator tokens failed %+v", err)
	}

	if !tokens.Success {
		log.Fatalf("Fetch authenticator tokens failed %+v", tokens)
	}

	mainpwd := d.getMainPassword()

	tks := []*Token{}
	for _, v := range tokens.AuthenticatorTokens {
		secret, err := v.Decrypt(mainpwd)

		if err != nil {
			log.Printf("Decryption failed for [%s]: %v", v.Name, err)
			continue
		}

		tks = append(tks, &Token{
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

		tks = append(tks, &Token{
			Name:    v.Name,
			Digital: v.Digits,
			Secret:  secret,
			Period:  10,
		})
	}

	d.tokenMap = tokensToMap(tks)
	d.tokens = tks
	d.saveToken()
	return
}

func (d *Device) getMainPassword() string {
	fmt.Println("d.registration.MainPassword", d.registration.MainPassword)
	if len(d.registration.MainPassword) == 0 {
		fmt.Print("\nPlease input Authy main password: ")
		pp, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Get password failed %+v", err)
		}

		d.registration.MainPassword = strings.TrimSpace(string(pp))
		d.SaveDeviceInfo()
	}

	return d.registration.MainPassword
}

func generateMD5(tk *Token) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(tk.Name+tk.OriginalName+tk.Secret)))
}

func tokensToMap(tks []*Token) map[string]*Token {
	ret := map[string]*Token{}
	for _, tk := range tks {
		ret[generateMD5(tk)] = tk
	}

	return ret
}

func (d *Device) saveToken() {
	regrPath, err := d.ConfigPath(cacheFileName)
	if err != nil {
		return
	}

	f, err := os.OpenFile(regrPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return
	}

	defer f.Close()

	tokens := make([]Token, 0, len(d.tokenMap))
	for _, v := range d.tokenMap {
		tokens = append(tokens, *v)
	}
	err = json.NewEncoder(f).Encode(tokens)
	return
}
