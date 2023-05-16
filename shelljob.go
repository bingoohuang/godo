package main

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/gg/pkg/ss"
)

var (
	envRe = regexp.MustCompile(`(?i)^export\s+([\w_-]+)=`)
	varRe = regexp.MustCompile(`\{(\d+),(\d+),(\d+)}`)
)

type Job struct {
	Shell string
	Vars  []Iterator
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
	for _, v := range j.Vars {
		jv := v.(*JobExpandVar)
		shell = strings.ReplaceAll(shell, jv.Var, jv.Value())
	}

	log.Printf("start to run shell %q", shell)
	cmd := exec.Command("bash", "-c", shell)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("cmd.Run %s failed with %s", j.Shell, err)
	}
}

func (j *Job) HasVar() bool { return len(j.Vars) > 0 }

type JobExpandVar struct {
	Var  string
	Curr int
	From int
	To   int
	Step int
}

func (v *JobExpandVar) HasNext() bool { return v.Curr <= v.To }
func (v *JobExpandVar) Next()         { v.Curr += v.Step }
func (v *JobExpandVar) Reset()        { v.Curr = v.From }
func (v *JobExpandVar) Value() string { return strconv.Itoa(v.Curr) }

func executeJobShell(shellData string) {
	env, jobGroups := parseJobsConfig(shellData)
	checkGroup(jobGroups)

	var interval time.Duration
	if gap := env["gap"]; gap != "" {
		interval, _ = time.ParseDuration(gap)
	}

	for _, group := range jobGroups {
		group.run(interval)
	}
}

func checkGroup(groups []jobGroup) {
	for _, group := range groups {
		group.Check()
	}
}

func (g jobGroup) run(interval time.Duration) {
	if g.VarJobIndex < 0 {
		var wg sync.WaitGroup
		for _, job := range g.Jobs {
			job.Run(&wg)
		}
		wg.Wait()
		time.Sleep(interval)
		return
	}

	out := make(chan struct{})
	wait := make(chan struct{})
	go Arrange(g.Jobs[g.VarJobIndex].Vars, out, wait)

	for range out {
		var wg sync.WaitGroup
		for _, job := range g.Jobs {
			job.Run(&wg)
		}
		wg.Wait()

		wait <- struct{}{}
		time.Sleep(interval)
	}
}

func (g *jobGroup) Check() {
	g.VarJobIndex = -1
	count := 0
	for i, job := range g.Jobs {
		if job.HasVar() {
			count++
			g.VarJobIndex = i
		}
	}

	if count > 1 {
		log.Fatalf("there are multiple jobs which has vars, only one for a group")
	}
}

type jobGroup struct {
	Jobs        []*Job
	VarJobIndex int
}

func parseJobsConfig(shellData string) (map[string]string, []jobGroup) {
	lines := ss.Split(shellData, ss.WithSeps("\n"), ss.WithTrimSpace(true), ss.WithIgnoreEmpty(true))
	var group jobGroup
	var groups []jobGroup
	env := map[string]string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "###") {
			if len(group.Jobs) > 0 {
				groups = append(groups, group)
			}
			group.Jobs = make([]*Job, 0)
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

		group.Jobs = append(group.Jobs, parseShellJob(line))
	}

	if len(group.Jobs) > 0 {
		groups = append(groups, group)
	}

	return env, groups
}

func parseShellJob(line string) *Job {
	job := Job{
		Shell: line,
	}

	for _, sub := range varRe.FindAllStringSubmatch(line, -1) {
		v := &JobExpandVar{
			Var:  sub[0],
			From: ss.ParseInt(sub[1]),
			To:   ss.ParseInt(sub[2]),
			Step: ss.ParseInt(sub[3]),
		}
		v.Reset()
		job.Vars = append(job.Vars, v)
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

// Arrangement 给出 n 维状态的组合列表.
// sss [n][m], 为 n 维状态，每维有 m 个状态
// e.g.  {{"红","黄","蓝"},{"方","圆"}}
// Arrangement 将返回：{{"红","方"},{"红","圆"}, {"黄","方"},{"黄","圆"}, {"蓝","方"}{"蓝","圆"}}，共6种组合状态
func Arrangement(sss [][]string) (ret [][]string) {
	n := len(sss)
	if n == 0 { // 0 维，0 种组合状态
		return nil
	} else if n == 1 { // 1 维， m 种组合状态，递归终止条件
		for _, li := range sss[0] {
			ret = append(ret, []string{li})
		}
		return ret
	}

	left := Arrangement(sss[:n-1]) // 取左边 n-1 个维
	right := sss[n-1]              // 取右边1个维

	// 递归，组合左右 2 个维状态
	for _, x := range left {
		for _, y := range right {
			reti := append([]string{}, x...)
			ret = append(ret, append(reti, y))
		}
	}

	return ret
}
