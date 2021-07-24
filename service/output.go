package service

import (
	"encoding/json"
	"fmt"
)

// AlfredOutput ..
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

// Output cale token output
type Output struct {
	*Token
	OTitle     string
	Code       string
	RemainSecs int

	Error error
}

// AfredSubtitle alfred subtitle
func (o Output) AfredSubtitle() string {
	if o.Error != nil {
		return o.Error.Error()
	}

	return fmt.Sprintf("Code: %s [Press Enter copy to clipboard], Expires in %d second(s)", o.Code, o.RemainSecs)
}

// Title ...
func (o Output) Title() string {
	if o.Token != nil {
		return o.Token.Title()
	}

	return o.OTitle
}

// ToAfred  to alfred output
func (o Output) ToAfred() AlfredOutput {
	return AlfredOutput{
		Title:    o.Title(),
		Subtitle: o.AfredSubtitle(),
		Arg:      o.Code,
		Valid:    true,
	}
}

func (s *Searcher) showResult(outputs []Output) {
	if s.isAlfred {
		s.showAlfredOutput(outputs)
		return
	}
	//	s.prettyPrintResult(outputs)
}

func (s *Searcher) showAlfredOutput(out []Output) {
	alfredOut := make([]AlfredOutput, 0, len(out))
	for _, v := range out {
		alfredOut = append(alfredOut, v.ToAfred())
	}

	m := map[string][]AlfredOutput{"items": alfredOut}
	b, _ := json.Marshal(m)
	fmt.Println(string(b))
}

const (
	// Red red
	Red = "\033[1;31m%s\033[0m"
	// Green green
	Green = "\033[1;32m%s\033[0m"
	// Teal teal
	Teal = "\033[1;36m%s\033[0m"
)

func (s *Searcher) prettyPrintResult(outputs []Output) {
	fmt.Printf("\n")
	for _, tk := range outputs {
		fmt.Printf("- Title: "+Green+"\n", tk.Title())
		if tk.Error != nil {
			fmt.Printf("- %v\n\n", tk.Error)
		} else {
			fmt.Printf("- Code: "+Teal+" Expires in "+Red+"(s)\n\n", tk.Code, fmt.Sprint(tk.RemainSecs))
		}
	}

	return
}
