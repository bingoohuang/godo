package main

import (
	"flag"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/gou/ran"
	"github.com/sirupsen/logrus"
)

func main() {
	span := ""
	nums := ""
	shell := ""
	setup := ""

	flag.StringVar(&setup, "setup", "100", "setup num")
	flag.StringVar(&span, "span", "10m", "time span to do sth, eg. 1h, 10m")
	flag.StringVar(&nums, "nums", "1", "numbers range, eg 1-3")
	flag.StringVar(&shell, "shell", "", "shell to invoke")

	flag.Parse()

	spanDuration, err := time.ParseDuration(span)
	if err != nil {
		panic(err)
	}

	expandNums := ExpandRange(nums)
	expandNumsLen := uint64(len(expandNums))

	if expandNumsLen == 0 {
		logrus.Fatal("nums", nums, "is illegal")
	}

	if shell == "" {
		logrus.Fatal("shell required")
	}

	if setup != "" {
		logrus.Infof("start to setup sth")

		do(shell, setup)

		logrus.Infof("complete to setup sth")
	}

	for {
		logrus.Infof("start to do sth")

		randNum := expandNums[ran.IntN(expandNumsLen)]

		do(shell, randNum)

		logrus.Infof("start to sleep %s", span)
		time.Sleep(spanDuration)
	}
}

func do(shell string, randNum string) {
	cmd := exec.Command("sh", "-c", shell)
	cmd.Env = []string{"GODO_NUM=" + randNum}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorf("cmd.Run %s failed with %s\n", shell, err)
	}
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
