package survey

import (
	"fmt"
	"regexp"

	"github.com/alecaivazis/survey/core"
	"github.com/chzyer/readline"
	ansi "github.com/k0kubun/go-ansi"
)

// Confirm is a regular text input that accept yes/no answers.
type Confirm struct {
	Message string
	Default bool
}

// data available to the templates when processing
type confirmTemplateData struct {
	Confirm
	Answer string
}

// Templates with Color formatting. See Documentation: https://github.com/mgutz/ansi#style-format
var confirmQuestionTemplate = `
{{- color "green+hb"}}? {{color "reset"}}
{{- color "default+hb"}}{{ .Message }} {{color "reset"}}
{{- if .Answer}}
  {{- color "cyan"}}{{.Answer}}{{color "reset"}}
{{- else }}
  {{- color "white"}}{{if .Default}}(Y/n) {{else}}(y/N) {{end}}{{color "reset"}}
{{- end}}`

// the regex for answers
var (
	yesRx = regexp.MustCompile("^(?i:y(?:es)?)$")
	noRx  = regexp.MustCompile("^(?i:n(?:o)?)$")
)

func yesNo(t bool) string {
	if t {
		return "Yes"
	}
	return "No"
}

func (c *Confirm) getBool(rl *readline.Instance) (bool, error) {
	// start waiting for input
	val, err := rl.Readline()
	// if something went wrong
	if err != nil {
		// use the default value and bubble up
		return c.Default, err
	}

	// get the answer that matches the
	var answer bool
	switch {
	case yesRx.Match([]byte(val)):
		answer = true
	case noRx.Match([]byte(val)):
		answer = false
	case val == "":
		answer = c.Default
	default:
		// we didnt get a valid answer, so print error and prompt again
		out, err := core.RunTemplate(
			errorTemplate, fmt.Errorf("%q is not a valid answer, please try again.", val),
		)
		// if something went wrong
		if err != nil {
			// use the default value and bubble up
			return c.Default, err
		}
		// send the message to the user
		ansi.Print(out)

		answer, err = c.getBool(rl)
		// if something went wrong
		if err != nil {
			// use the default value
			return c.Default, err
		}
	}

	return answer, nil
}

// Prompt prompts the user with a simple text field and expects a reply followed
// by a carriage return.
func (c *Confirm) Prompt(rl *readline.Instance) (string, error) {
	// render the question template
	out, err := core.RunTemplate(
		confirmQuestionTemplate,
		confirmTemplateData{Confirm: *c},
	)
	if err != nil {
		return "", err
	}

	// use the result of the template as the prompt for the readline instance
	rl.SetPrompt(fmt.Sprintf(out))

	// start waiting for input
	answer, err := c.getBool(rl)
	// if something went wrong
	if err != nil {
		// bubble up
		return "", err
	}

	// convert the boolean into the appropriate string
	return yesNo(answer), nil
}

// Cleanup overwrite the line with the finalized formatted version
func (c *Confirm) Cleanup(rl *readline.Instance, val string) error {
	// go up one line
	ansi.CursorPreviousLine(1)
	// clear the line
	ansi.EraseInLine(1)

	// render the template
	out, err := core.RunTemplate(
		confirmQuestionTemplate,
		confirmTemplateData{Confirm: *c, Answer: val},
	)
	if err != nil {
		return err
	}

	// print the summary
	ansi.Println(out)

	// we're done
	return nil
}
