package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	Black  = `0`
	Red    = `1`
	Green  = `2`
	Brown  = `3`
	Blue   = `4`
	Purple = `5`
	Cyan   = `6`
	Gray   = `7`
)

// Color returns the ANSI colour code for the given background and foreground.
// Note Bold is usually interpeted as 'light' these days. E.g. 'light blue.'
func Color(bg string, fg string, bold bool) string {
	var boldstr string
	if bold {
		boldstr = `1`
	} else {
		boldstr = `0`
	}
	return "\033[4" + bg + `;` + boldstr + `;3` + fg + `m`

}

// InverseColor returns the inverse ANSI colour code, which, when sent to the terminal, will invert the current colours.
func InverseColor() string {
	return "\033[7m"
}

// ResetColor returns the reset ANSI colour code, which will reset the terminal colours to their default
func ResetColor() string {
	return "\033[0m"
}

// TerminalSize returns the width and height of the tty, respectively
func TerminalSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	outarr := strings.Split(string(out), " ")
	if len(outarr) != 2 {
		return 0, 0, errors.New("Terminal size split error, stty size return unexpected: " + string(out))
	}

	height, err := strconv.Atoi(outarr[0])
	if err != nil {
		return 0, 0, errors.New("Termianl size height error, stty size return unexpected: " + string(outarr[0]))
	}

	width, err := strconv.Atoi(strings.Trim(outarr[1], " \n"))
	if err != nil {
		return 0, 0, errors.New("Terminal size width error, stty size return unexpected: " + string(outarr[1]))
	}

	return width, height, nil
}

// JsonHeading represents a cheatsheet heading definition. This is designed to be parsed from JSON.
type JsonHeading struct {
	Title    string   `json:"title"`
	Commands []string `json:"commands"`
}

// JsonHeading represents a cheatsheet definition, of all headings. This is designed to be parsed from JSON.
type JsonHeadings struct {
	Headings []JsonHeading `json:"headings"`
}

// Width returns the width of this JsonHeading, which is the width of the longest command.
func (h JsonHeading) Width() int {
	max := len(h.Title)
	for _, v := range h.Commands {
		if len(v) > max {
			max = len(v)
		}
	}
	return max
}

// HeadingWidth returns the width which headings should be, which is the width of the widest heading
func (hs JsonHeadings) HeadingWidth() int {
	max := 0
	for _, v := range hs.Headings {
		if v.Width() > max {
			max = v.Width()
		}
	}
	return max
}

// CommandsHeight is the height which commands should be, which is the number of commands in the heading with the most commands.
func (hs JsonHeadings) CommandsHeight() int {
	max := 0
	for _, v := range hs.Headings {
		if len(v.Commands) > max {
			max = len(v.Commands)
		}
	}
	return max
}

// RowHasCommands returns whether a given row has any commands.
// This is used to not display blank lines, if all the headings in a given row have fewer than the tallest row.
func (hs JsonHeadings) RowHasCommands(start, end, row int) bool {
	for i := start; i < end; i++ {
		heading := hs.Headings[i]
		if len(heading.Commands) >= (row + 1) {
			return true
		}
	}
	return false
}

// PrintHeadings returns the heading string to be printed to the console, including ANSI color codes
func (hs *JsonHeadings) PrintHeadings(start, end, width int) string {
	if end > len(hs.Headings) {
		end = len(hs.Headings)
	}

	var s string
	for i := start; i != end; i++ {
		heading := hs.Headings[i]
		s += InverseColor()
		s += heading.Title + strings.Repeat(" ", width-len(heading.Title)-1)
		s += ResetColor()
		s += ` `
	}

	commandsHeight := hs.CommandsHeight()
	for j := 0; j != commandsHeight; j++ {
		if !hs.RowHasCommands(start, end, j) {
			continue
		}
		s += "\n"
		for i := start; i < end; i++ {
			heading := hs.Headings[i]
			if len(heading.Commands) <= j {
				s += strings.Repeat(" ", width)
				continue
			}
			command := heading.Commands[j]
			s += command + strings.Repeat(" ", width-len(command))
		}
	}

	return s
}

// PrintString returns a string for printing, of the headings, including ANSI color codes, formatted to the given width.
func (hs *JsonHeadings) PrintString(width int) string {
	var s string

	headingWidth := hs.HeadingWidth() + 1
	headingsPerLine := width / headingWidth // +1 because headings are separated
	for i := 0; i < len(hs.Headings); i += headingsPerLine {
		s += hs.PrintHeadings(i, i+headingsPerLine, headingWidth) + "\n"
	}
	return s
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`Usage: cman <program>`)
		return
	}

	filename := os.Args[1] + `.json`

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println(`Cheatsheet Man does not exist for ` + os.Args[1])
		return
	}

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		return
	}

	var jsonHeadings JsonHeadings
	err = json.Unmarshal(file, &jsonHeadings)
	if err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		return
	}

	width, _, err := TerminalSize()
	if err != nil {
		fmt.Printf("Terminal size error: %v\n", err)
		return
	}

	fmt.Println(jsonHeadings.PrintString(width))
}
