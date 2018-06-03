package spent

import (
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type (
	// Reporter produces activity reports based on observations of focused
	// window title.
	Reporter struct {
		// lastActive is what we observed as the focused window last time we looked.
		lastActive string
		// lastReport is when we last produced a non-empty report.
		lastReport time.Time
		// writeInterval is how often a report should be produced in the absence of change.
		writeInterval time.Duration
	}

	Report struct {
		At         time.Time
		Elapsed    time.Duration
		Title      string
		App        string
		AppContext string
		AppDetail  string
	}
)

// NewReporter returns a new Reporter.
func NewReporter(writeInterval time.Duration) *Reporter {
	return &Reporter{
		lastActive:    "",
		lastReport:    time.Now(),
		writeInterval: writeInterval,
	}
}

func newReport(at time.Time, title string, elapsed time.Duration) *Report {
	rpt := &Report{At: at, Title: title, Elapsed: elapsed}
	rpt.extractAppFields()
	return rpt
}

func (rpt *Report) extractAppFields() {
	appfields := parseTitle(rpt.Title)
	if len(appfields) > 2 {
		rpt.AppDetail = appfields[2]
	}
	if len(appfields) > 1 {
		rpt.AppContext = appfields[1]
	}
	if len(appfields) > 0 {
		rpt.App = appfields[0]
	}
}

// GetReport returns nil if there's nothing to report, otherwise a string
// slice describing what's happened.  See README.md for details of slice
// contents.
func (r *Reporter) GetReport(title string) *Report {
	now := time.Now()
	delta := now.Sub(r.lastReport)
	if title == r.lastActive && delta < r.writeInterval {
		return nil
	}

	var result *Report
	if r.lastActive != "" {
		result = newReport(now, r.lastActive, delta)
	}
	r.lastReport, r.lastActive = now, title
	return result
}

func parseBrowserTitle(s string) []string {
	i := strings.LastIndex(s, " - ")
	if i == -1 {
		return []string{}
	}

	u, err := url.Parse(s[i+3:])
	if err != nil {
		return []string{}
	}
	return []string{u.Host, u.Path}
}

func parseTerminalTitle(s string) []string {
	i := strings.IndexByte(s, ':')
	if i == -1 {
		return []string{}
	}
	return []string{s[:i], strings.TrimSpace(s[i+1:])}
}

func parseEditorTitle(s string) []string {
	i := strings.LastIndex(s, " - ")
	return []string{s[i+3:], s[:i]}
}

func parseTitleApp(title string) (string, string) {
	i := strings.LastIndex(title, " - ")
	if i == -1 || i == len(title)-1 {
		return "", ""
	}
	first, rest := strings.TrimSpace(title[:i]), strings.TrimSpace(title[i+3:])
	if first == "" || rest == "" {
		return "", ""
	}
	return first, rest
}

// parseTitle examines a window title and returns its components.
func parseTitle(title string) []string {
	first, rest := parseTitleApp(title)

	switch rest {
	case "Chromium":
		return append([]string{"browser"}, parseBrowserTitle(first)...)
	case "Visual Studio Code":
		return append([]string{"editor"}, parseEditorTitle(first)...)
	case "xterm":
		return append([]string{"terminal"}, parseTerminalTitle(first)...)
	}

	return []string{}
}

func GetIdleTime() (time.Duration, error) {
	cmd := exec.Command("xprintidle")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0 * time.Second, err
	}

	ms, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0 * time.Second, err
	}

	return time.Duration(ms) * time.Millisecond, nil
}

func GetActiveWindow() (string, error) {
	cmd := exec.Command("xdotool", "getwindowfocus", "getwindowname")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func GetScreensaverOn() (bool, error) {
	cmd := exec.Command("gnome-screensaver-command", "-q")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	switch strings.TrimSpace(string(out)) {
	case "The screensaver is inactive":
		return false, nil
	case "The screensaver is active":
		return true, nil
	default:
		return false, fmt.Errorf("Unexpected output from gnome-screensaver-command -q: %v", out)
	}

}
