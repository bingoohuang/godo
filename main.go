package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/bingoohuang/gg/pkg/ctl"
	"github.com/bingoohuang/gg/pkg/fla9"
	"github.com/bingoohuang/gg/pkg/randx"
	"github.com/bingoohuang/gg/pkg/thinktime"
	"github.com/bingoohuang/godaemon"
	"github.com/bingoohuang/golog"
)

func main() {
	log.SetOutput(os.Stdout)

	var app App

	app.ParseFlags()

	var wg sync.WaitGroup
	if app.shell != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.runShell()
		}()
	}

	if app.conf != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.runConf()
		}()
	}

	wg.Wait()
}

func (a *App) runShell() {
	data, err := os.ReadFile(a.shell)
	if err != nil {
		log.Printf("read shell %s: %v", a.shell, err)
		return
	}

	if bytes.HasPrefix(data, []byte("###")) {
		executeJobShell(string(data))
		return
	}

	a.SetupJob()
	a.LoopJob()
	return
}

// App structures the godo application.
type App struct {
	think   *thinktime.ThinkTime
	shell   string
	conf    string
	setup   string
	nums    []string
	env     []string
	numsLen uint64
}

// ParseFlags parses the command line arguments.
func (a *App) ParseFlags() {
	pInit := fla9.Bool("init", false, "Create initial ctl and exit")
	pVersion := fla9.Bool("version,v", false, "Create initial ctl and exit")
	pDeamon := fla9.Bool("deamon,d", false, "Run in daemonize mode")
	fla9.StringVar(&a.setup, "setup", "", "setup num")
	span := fla9.String("span", "10s", "time span to do sth, eg. 1h, 10m for fixed span, "+
		"or 10s-1m for rand span among the range")
	nums := fla9.String("nums", "1", "numbers range, eg 1-3")
	fla9.StringVar(&a.shell, "shell", "", "shell to invoke")
	fla9.StringVar(&a.conf, "conf,c", "", "yaml")
	fla9.Parse()

	ctl.Config{Initing: *pInit, PrintVersion: *pVersion}.ProcessInit()
	golog.Setup()

	godaemon.Daemonize(godaemon.WithDaemon(*pDeamon))

	for _, arg := range fla9.Args() {
		if name, value, ok := splitNameValue(arg); ok {
			a.env = append(a.env, name+"="+value)
		}
	}

	log.Printf("env: %v", a.env)

	a.parseSpan(*span)

	a.nums = ExpandRange(*nums)
	a.numsLen = uint64(len(a.nums))

	if a.numsLen == 0 {
		log.Fatal("nums", nums, "is illegal")
	}

	if a.shell == "" && a.conf == "" {
		log.Fatal("shell/conf required")
	}
}

func splitNameValue(arg string) (name, value string, ok bool) {
	pos := strings.IndexAny(arg, "=:")
	if pos <= 0 {
		return "", "", false
	}

	name = strings.TrimSpace(arg[:pos])
	if pos < len(arg) {
		value = strings.TrimSpace(arg[pos+1:])
	}
	return name, value, true
}

func (a *App) parseSpan(span string) {
	var err error
	a.think, err = thinktime.ParseThinkTime(span)
	if err != nil {
		log.Fatal(err)
	}
}

// SetupJob setup job before looping.
func (a *App) SetupJob() {
	if a.setup == "" {
		log.Printf("nothing to setup")
		return
	}

	log.Printf("start to setup sth")
	a.executeShell(a.shell, a.setup)
	log.Printf("complete to setup sth")
}

func (a *App) executeShell(shell, randNum string) {
	var env []string
	env = append(env, os.Environ()...)
	env = append(env, a.env...)
	env = append(env, "GODO_NUM="+randNum)

	cmd := exec.Command("sh", "-c", shell)
	cmd.Env = env
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("cmd.Run %s: %v", shell, err)
	} else {
		sout := strings.TrimFunc(string(output), unicode.IsSpace)
		log.Printf("[PRE]Do job, got %s", sout)
	}
}

// LoopJob loop job in an infinite loop.
func (a *App) LoopJob() {
	for {
		log.Printf("start to do sth")
		go a.executeShell(a.shell, a.nums[randx.IntN(int(a.numsLen))])
		a.randSpan()
	}
}

func (a *App) randSpan() {
	if a.think != nil {
		a.think.Think(true)
	}
	a.stopCheck()
}

func (a *App) stopCheck() {
	var env []string
	env = append(env, os.Environ()...)
	env = append(env, a.env...)
	env = append(env, "GODO_NUM=exitCheck")

	cmd := exec.Command("sh", "-c", a.shell)
	cmd.Env = env

	output, _ := cmd.CombinedOutput()
	sout := strings.TrimFunc(string(output), unicode.IsSpace)
	if !strings.EqualFold(sout, "exit") {
		log.Printf("[PRE] Stop check, got %s, continue", sout)
		return
	}

	log.Printf("[PRE] Stop check, got %s, exiting", sout)
	os.Exit(0)
}

func (a *App) runConf() {
	conf, err := ParseConf(a.conf)
	if err != nil {
		log.Printf("parse %s: %v", a.conf, err)
		return
	}

	conf.Run()
}

// ExpandRange expands a string like 1-3 to [1,2,3]
func ExpandRange(f string) []string {
	hyphenPos := strings.Index(f, "-")
	if hyphenPos <= 0 || hyphenPos == len(f)-1 {
		return []string{f}
	}

	from := strings.TrimSpace(f[0:hyphenPos])
	to := strings.TrimSpace(f[hyphenPos+1:])

	fromI := 0
	toI := 0

	var err error

	if fromI, err = strconv.Atoi(from); err != nil {
		return []string{f}
	}

	if toI, err = strconv.Atoi(to); err != nil {
		return []string{f}
	}

	parts := make([]string, 0)

	if fromI < toI {
		for i := fromI; i <= toI; i++ {
			parts = append(parts, strconv.Itoa(i))
		}
	} else {
		for i := fromI; i >= toI; i-- {
			parts = append(parts, strconv.Itoa(i))
		}
	}

	return parts
}
