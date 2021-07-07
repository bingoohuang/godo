package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/gou/ran"
)

func main() {
	var app App

	app.ParseFlags()
	app.SetupJob()
	app.LoopJob()
}

// App structures the godo application.
type App struct {
	startSpan time.Duration
	endSpan   time.Duration
	nums      []string
	shell     string
	setup     string
	numsLen   uint64
}

// ParseFlags parses the command line arguments.
func (a *App) ParseFlags() {
	flag.StringVar(&a.setup, "setup", "", "setup num")
	span := flag.String("span", "10m", "time span to do sth, eg. 1h, 10m for fixed span, "+
		"or 10s-1m for rand span among the range")
	nums := flag.String("nums", "1", "numbers range, eg 1-3")
	flag.StringVar(&a.shell, "shell", "", "shell to invoke")

	flag.Parse()

	a.parseSpan(*span)

	a.nums = ExpandRange(*nums)
	a.numsLen = uint64(len(a.nums))

	if a.numsLen == 0 {
		log.Fatal("nums", nums, "is illegal")
	}

	if a.shell == "" {
		log.Fatal("shell required")
	}
}

func (a *App) parseSpan(span string) {
	var err error

	if strings.Contains(span, "-") {
		pos := strings.Index(span, "-")

		if a.startSpan, err = time.ParseDuration((span)[:pos]); err != nil {
			panic(err)
		}

		if a.endSpan, err = time.ParseDuration((span)[pos+1:]); err != nil {
			panic(err)
		}
	} else if a.startSpan, err = time.ParseDuration(span); err != nil {
		panic(err)
	}
}

// SetupJob setup job before looping.
func (a *App) SetupJob() {
	if a.setup == "" {
		log.Printf("nothing to setup")
	}

	log.Printf("start to setup sth")

	a.executeShell(a.shell, a.setup)

	log.Printf("complete to setup sth")
}

func (a *App) executeShell(shell, randNum string) {
	cmd := exec.Command("sh", "-c", shell)
	cmd.Env = []string{"GODO_NUM=" + randNum}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("cmd.Run %s failed with %s", shell, err)
	}
}

// LoopJob loop job in an infinite loop.
func (a *App) LoopJob() {
	for {
		log.Printf("start to do sth")

		randNum := a.nums[ran.IntN(a.numsLen)]

		go a.executeShell(a.shell, randNum)

		span := a.randSpan()
		log.Printf("start to sleep %s", span)
		time.Sleep(span)
	}
}

func (a *App) randSpan() time.Duration {
	if a.endSpan == 0 {
		return a.startSpan
	}

	return a.startSpan + time.Duration(ran.IntN(uint64(a.endSpan-a.startSpan)))
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
