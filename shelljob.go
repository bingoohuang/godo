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

func (j *Job) Complete() bool {
	return j.Var == nil || j.Var.Complete()
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
	From int
	To   int
	Step int
}

func (v *JobExpandVar) Complete() bool {
	return v.From > v.To
}

func (v *JobExpandVar) Value() string {
	from := v.From
	v.From += v.Step
	return strconv.Itoa(from)
}

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
	complete := false
	for !complete {
		var wg sync.WaitGroup
		for _, job := range jobs {
			job.Run(&wg)
		}
		wg.Wait()

		complete = completeJobs(jobs)
	}
}

func completeJobs(jobs []*Job) bool {
	for _, job := range jobs {
		if !job.Complete() {
			return false
		}
	}

	return true
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
	}

	return &job
}
