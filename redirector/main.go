package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

type RedirectHandler func([]string) (string, error)

type Redirector struct {
	Pattern string
	Handler RedirectHandler
}

var redirectors = []Redirector{
	Redirector{`/r([0-9a-fA-F]{6,40})/?`, HandleGitRevision},
	Redirector{`/pr(\d+)/?`, HandlePullRequest},

	Redirector{`/pr$`, MakeStaticRedirector("https://github.com/dolphin-emu/dolphin/pulls")},
	Redirector{`/dl$`, MakeStaticRedirector("https://dolphin-emu.org/download/")},
	Redirector{`/gh$`, MakeStaticRedirector("https://github.com/dolphin-emu/dolphin")},
	Redirector{`/faq$`, MakeStaticRedirector("https://dolphin-emu.org/docs/faq/")},
	Redirector{`/bbs$`, MakeStaticRedirector("https://forums.dolphin-emu.org/")},
}

func MakeStaticRedirector(url string) RedirectHandler {
	return func ([]string) (string, error) {
		return url, nil
	}
}

func HandleGitRevision(args []string) (string, error) {
	return fmt.Sprintf("https://github.com/dolphin-emu/dolphin/commit/%s", args[0]), nil
}

func HandlePullRequest(args []string) (string, error) {
	return fmt.Sprintf("https://github.com/dolphin-emu/dolphin/pull/%s", args[0]), nil
}

var readmeContents = GetReadme()

func GetReadme() string {
	s, err := ioutil.ReadFile("README")
	if err != nil {
		return "No README found. Cannot provide you with documentation ☹"
	}
	return string(s)
}

func Router(w http.ResponseWriter, req *http.Request) {
	for _, r := range redirectors {
		re := regexp.MustCompile("^" + r.Pattern + "$")
		matches := re.FindStringSubmatch(req.URL.Path)
		if matches != nil {
			url, err := r.Handler(matches[1:])
			if err != nil {
				fmt.Fprintf(w, "Error: %v: %v", r.Handler, err)
				return
			}
			http.Redirect(w, req, url, 302)
			return
		}
	}

	fmt.Fprintf(w, readmeContents)
}

func main() {
	http.HandleFunc("/", Router)
	http.ListenAndServe(":8033", nil)
}
