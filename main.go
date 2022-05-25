package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"golang.org/x/xerrors"
)

type pr struct {
	Number      string
	MergeCommit string
}

type tmplVar struct {
	PRs    string
	Branch string
}

var (
	dryRun        = flag.Bool("dry-run", false, "dry run")
	iprs          = flag.String("prs", "", "comma separated pr numbers")
	toBranch      = flag.String("to", "", "to branch name that you want to pick PRs")
	branchPrefix  = flag.String("branch-prefix", "pick-", "branch name prefix for created PR")
	prTitlePrefix = flag.String("title-prefix", "[pick]", "title prefix for created PR")
	bodyTmpl      = flag.String("body", "This PR is picking {{ .PRs }} to {{ .Branch }} branch.", "body template by go text/template. you can use .PRs, .Branch variable.")
	workDir       = flag.String("workdir", "", "working directory")
)

func main() {
	flag.Parse()
	if *iprs == "" {
		log.Fatal("need -prs")
	}
	if *toBranch == "" {
		log.Fatal("need -to")
	}
	if *workDir != "" {
		if err := os.Chdir(*workDir); err != nil {
			log.Fatal(err)
		}
	}
	prNums := strings.Split(*iprs, ",")

	if _, err := exec.LookPath("gh"); err != nil {
		log.Fatalf("please install gh command. see https://github.com/cli/cli: %v", err)
	}
	if err := exec.Command("gh", "auth", "status").Run(); err != nil {
		log.Fatalf("you need `gh auth login` to use logined gh command: %v", err)
	}
	if err := exec.Command("git", "fetch", "origin").Run(); err != nil {
		log.Fatalf("git fetch origin failed: %v", err)
	}

	prs, err := getPRs(prNums)
	if err != nil {
		log.Fatalf("cant get PRs from main commit log :%v", err)
	}

	var (
		newBranch string
		title     string
		b         strings.Builder
		s         tmplVar
	)

	t, err := template.New("tmpl").Parse(*bodyTmpl)
	if err != nil {
		log.Fatal(err)
	}
	s.Branch = *toBranch
	if len(prNums) == 1 {
		newBranch = fmt.Sprintf("%s%s", *branchPrefix, prNums[0])
		title = fmt.Sprintf("\"%s #%s\"", *prTitlePrefix, prNums[0])
		s.PRs = fmt.Sprintf("#%s", prNums[0])
	} else {
		newBranch = fmt.Sprintf("%s%s-to-%s", *branchPrefix, prNums[0], prNums[len(prNums)-1])
		title = fmt.Sprintf("\"%s %s\"", *prTitlePrefix, strings.Join(hashed(prNums), ", "))
		s.PRs = strings.Join(hashed(prNums), ", ")
	}
	if err = t.Execute(&b, s); err != nil {
		log.Fatal(err)
	}

	printOrExec(fmt.Sprintf("git checkout --no-track -b %s origin/%s", newBranch, *toBranch), *dryRun)
	for _, pr := range prs {
		printOrExec(fmt.Sprintf("git cherry-pick -m 1 %s", pr.MergeCommit), *dryRun)
	}
	printOrExec(fmt.Sprintf("git push origin %s", newBranch), *dryRun)
	printOrExec(fmt.Sprintf("gh pr create -B %s --title %s --body %s", *toBranch, title, b.String()), *dryRun)

}

func getPRs(prNums []string) ([]pr, error) {
	prs := []pr{}
	out, err := exec.Command("sh", "-c", "git log --merges  --pretty=\"%h|%s\" |grep \"Merge pull request\"").Output()
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("cant get git log")
	}
	for _, prNum := range prNums {
		for _, log := range strings.Split(string(out), "\n") {
			cols := strings.Split(log, "|")
			if len(cols) < 2 {
				continue
			}
			if strings.HasPrefix(cols[1], fmt.Sprintf("Merge pull request #%s from", prNum)) {
				prs = append(prs, pr{
					Number:      prNum,
					MergeCommit: cols[0],
				})
				break
			}
		}
	}
	if len(prs) == len(prNums) {
		return prs, nil
	} else {
		return nil, xerrors.New("some pr does not found from git log.")
	}
}

func printOrExec(c string, dryRun bool) {
	fmt.Println(c)
	if !dryRun {
		o, err := exec.Command("sh", "-c", c).CombinedOutput()
		if err != nil {
			log.Fatalf("%s failed :%v", c, err)
		}
		fmt.Println(string(o))
	}
}

func hashed(in []string) []string {
	r := make([]string, len(in))
	for i, s := range in {
		r[i] = "#" + s
	}
	return r
}
