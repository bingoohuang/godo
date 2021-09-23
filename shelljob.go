package main

import (
	"github.com/bingoohuang/gg/pkg/ss"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var envRe = regexp.MustCompile(`(?i)^export\s+([\w\d_-]+)=`)
var varRe = regexp.MustCompile(`\{(\d+),(\d+),(\d+)}`)

type Job struct {
	Shell string
	Var   *JobExpandVar
}

func (j *Job) HasNext() bool {
	return j.Var != nil && j.Var.HasNext()
}

func (j *Job) Run(s *sync.WaitGroup) {
	s.Add(1)
	go func() {
		defer s.Done()
		j.goRun()
	}()
}

func (j *Job) goRun() {
	shell := j.Shell
	if j.Var != nil {
		shell = strings.ReplaceAll(shell, j.Var.Var, j.Var.Value())
	}

	log.Printf("start to run shell %q", shell)
	cmd := exec.Command("bash", "-c", shell)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("cmd.Run %s failed with %s", j.Shell, err)
	}
}

type JobExpandVar struct {
	Var  string
	Curr int
	From int
	To   int
	Step int
}

func (v *JobExpandVar) HasNext() bool { return v.Curr <= v.To }

func (v *JobExpandVar) Next() {
	if v.HasNext() {
		v.Curr += v.Step
	} else {
		v.Curr = v.From
	}
}

func (v *JobExpandVar) Value() string { return strconv.Itoa(v.Curr) }

func (v *JobExpandVar) Init() { v.Curr = v.From }

func executeJobShell(shellData string) {
	env, allJobs := parseJobsConfig(shellData)

	var interval time.Duration
	if gap := env["gap"]; gap != "" {
		interval, _ = time.ParseDuration(gap)
	}

	for _, jobs := range allJobs {
		runJobs(jobs)
		time.Sleep(interval)
	}
}

func runJobs(jobs []*Job) {
	for {
		var wg sync.WaitGroup
		for _, job := range jobs {
			job.Run(&wg)
		}
		wg.Wait()

		if !nextJobs(jobs) {
			break
		}
	}
}

func nextJobs(jobs []*Job) bool {
	for i := len(jobs) - 1; i >= 0; i-- {
		job := jobs[i]
		if job.HasNext() {
			return true
		}
	}

	return false
}

func parseJobsConfig(shellData string) (map[string]string, [][]*Job) {
	lines := ss.Split(shellData, ss.WithSeps("\n"), ss.WithTrimSpace(true), ss.WithIgnoreEmpty(true))
	var allJobs [][]*Job
	var jobs []*Job
	env := map[string]string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "###") {
			if len(jobs) > 0 {
				allJobs = append(allJobs, jobs)
			}
			jobs = make([]*Job, 0)
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if sub := envRe.FindStringSubmatch(line); len(sub) > 0 {
			env[sub[1]] = line[len(sub[0]):]
			continue
		}

		for k, v := range env {
			line = strings.ReplaceAll(line, `${`+k+`}`, v)
		}

		jobs = append(jobs, parseShellJob(line))
	}

	if len(jobs) > 0 {
		allJobs = append(allJobs, jobs)
	}
	return env, allJobs
}

func parseShellJob(line string) *Job {
	job := Job{
		Shell: line,
	}

	if vars := varRe.FindStringSubmatch(line); len(vars) > 0 {
		job.Var = &JobExpandVar{
			Var:  vars[0],
			From: ss.ParseInt(vars[1]),
			To:   ss.ParseInt(vars[2]),
			Step: ss.ParseInt(vars[3]),
		}
		job.Var.Init()
	}

	return &job
}

type Iterator interface {
	HasNext() bool
	Next()
	Reset()
}

func Arrange(sss []Iterator, out chan struct{}, wait chan struct{}) {
	defer close(out)

	l := len(sss)
	right := sss[l-1]
	if l == 1 {
		rotate(right, out, wait)
		return
	}

	leftOut := make(chan struct{})
	leftWait := make(chan struct{})
	go Arrange(sss[:l-1], leftOut, leftWait)
	for range leftOut {
		rotate(right, out, wait)
		leftWait <- struct{}{}
	}
}

func rotate(iter Iterator, out, wait chan struct{}) {
	for iter.Reset(); iter.HasNext(); iter.Next() {
		out <- struct{}{}
		<-wait
	}
}

func Arrangement(sss [][]string) (ret [][]string) {
	l := len(sss)
	if l == 0 {
		return nil
	} else if l == 1 {
		for _, li := range sss[0] {
			ret = append(ret, []string{li})
		}
		return ret
	}

	left := Arrangement(sss[:l-1])
	right := sss[l-1]

	for _, x := range left {
		for _, y := range right {
			reti := append([]string{}, x...)
			ret = append(ret, append(reti, y))
		}
	}

	return ret
}
