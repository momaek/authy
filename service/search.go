package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/momaek/authy/totp"
	"github.com/sahilm/fuzzy"
)

// Searcher
type Searcher struct {
	isAlfred bool
	keyword  string
	*Device
}

func (s *Searcher) showAll() bool {
	return len(s.keyword) == 0 || s.keyword == ""
}

// NewSearcher ..
func NewSearcher(keyword string, isAlfred bool) *Searcher {
	return &Searcher{
		isAlfred: isAlfred,
		keyword:  keyword,
		Device:   NewDevice(NewDeviceConfig{}),
	}
}

// Search fuzzy search tokens by name and origin_name
func (s *Searcher) Search() {
	var (
		tokens  = []*Token{}
		outputs = []Output{}
	)

	s.Device.LoadTokenFromCache()

	if s.showAll() {
		sort.Sort(Tokens(s.Device.tokens))
		tokens = s.Device.tokens
	} else {
		tokens = s.searchTokens()
	}

	if len(s.Device.tokens) == 0 {
		outputs = append(outputs, Output{
			OTitle: "OTP tokens not found",
			Error:  errors.New("Please run 'authy refresh' in commandline"),
		})
	} else {
		outputs = s.calcTokens(tokens)
	}

	s.showResult(outputs)
}

func (s *Searcher) searchTokens() []*Token {
	results := fuzzy.FindFrom(s.keyword, Tokens(s.Device.tokens))
	foundTokens := make([]*Token, 0, len(results))
	for _, v := range results {
		foundTokens = append(foundTokens, s.Device.tokens[v.Index])
		s.Device.tokens[v.Index].updateWeight()
	}
	s.Device.saveToken()
	return foundTokens
}

func calcRemainSec(challenge int64) int {
	return 30 - int(time.Now().Unix()-challenge*30)
}

func (s *Searcher) calcTokens(tokens []*Token) []Output {
	out := make([]Output, 0, len(tokens))
	for _, tk := range tokens {
		if len(tk.Secret) == 0 {
			out = append(out, Output{
				Token: tk,
				Error: errors.New("OTP token is empty"),
			})
			continue
		}

		codes := totp.GetTotpCode(tk.Secret, tk.Digital)
		challenge := totp.GetChallenge()

		out = append(out, Output{
			Token:      tk,
			Code:       codes[1],
			RemainSecs: calcRemainSec(challenge),
		})
	}

	if len(out) == 0 {
		out = append(out, Output{
			OTitle: fmt.Sprintf("OTP token not found (%s)", s.keyword),
			Error:  errors.New("Please try another keyword"),
		})
	}

	return out
}
