package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/rjeczalik/notify"
	"gopkg.in/yaml.v3"
)

// ParseConf 解析 YAML 配置文件
func ParseConf(filePath string) (*Godo, error) {
	cf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s failed: %w", filePath, err)
	}

	var c Godo

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(cf, &c); err != nil {
			return nil, fmt.Errorf("unmarshal yaml file %s failed: %w", filePath, err)
		}
		return &c, nil
	case ".json", ".js":
		if err := json.Unmarshal(cf, &c); err != nil {
			return nil, fmt.Errorf("unmarshal json file %s failed: %w", filePath, err)
		}
		return &c, nil
	}

	return nil, fmt.Errorf("unknown file format: %s", filePath)
}

type Godo struct {
	*Watch `json:"watch"`
}

func (g *Godo) Run() {
	if g.Watch != nil {
		g.Watch.run()
	}
}

type Watch struct {
	Directories []string `json:"directories"`
	Operations  []string `json:"operations"`
	ops         []notify.Event
	Actions     []*Action `json:"actions"`
}

type ExecuteAction interface {
	ExecuteAction(file string)
}

type Action struct {
	Matches []string `json:"matches"`
	Actions []string `json:"actions"`

	actions []ExecuteAction
}

func (w *Watch) Do(file string) {
	for _, action := range w.Actions {
		action.Do(file)
	}
}

func (w *Action) Do(file string) {
	if !w.matches(file) {
		return
	}

	for _, action := range w.actions {
		action.ExecuteAction(file)
	}
}

func (w *Action) matches(file string) bool {
	base := filepath.Base(file)
	for _, match := range w.Matches {
		if strings.HasPrefix(match, "!") {
			matched, err := filepath.Match(match[1:], base)
			if err != nil {
				log.Printf("bad match: %v", err)
				continue
			}
			if !matched {
				return true
			}
		} else {
			matched, err := filepath.Match(match, base)
			if err != nil {
				log.Printf("bad match: %v", err)
				continue
			}

			if matched {
				return true
			}
		}
	}

	return false
}

func (w *Action) postInit() {
	w.actions = parseActions(w.Actions)
}

func parseActions(actions []string) (parsed []ExecuteAction) {
	parsed = make([]ExecuteAction, 0, len(actions))

	for _, action := range actions {
		parsed = append(parsed, parseAction(action))
	}

	return parsed
}

type dotCleanAction struct{}

func (d dotCleanAction) ExecuteAction(file string) {
	dir := filepath.Dir(file)
	shell := "dot_clean " + strconv.Quote(dir)
	cmd := exec.Command("sh", "-c", shell)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("shell %s: %v", shell, err)
	} else {
		log.Printf("shell %s: %s", shell, ss.Or(string(output), "(Empty)"))
	}
}

type removeAction struct{}

func (r removeAction) ExecuteAction(file string) {
	if err := os.Remove(file); err != nil {
		log.Printf("remove %s: %v", file, err)
	} else {
		log.Printf("remove %s successfully", file)
	}
}

type noopAction struct{}

func (n noopAction) ExecuteAction(string) {}

func parseAction(action string) ExecuteAction {
	if strings.EqualFold(action, "remove") {
		return &removeAction{}
	}
	if strings.EqualFold(action, "dot_clean") {
		return &dotCleanAction{}
	}

	log.Printf("unknown action: %s", action)
	return &noopAction{}
}

func (w *Watch) postInit() {
	w.ops = parseOperations(w.Operations)
	for _, action := range w.Actions {
		action.postInit()
	}
}

func (w *Watch) run() {
	w.postInit()

	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 1)

	// Set up a watchpoint listening for events within a directory tree rooted
	// at current working directory. Dispatch remove events to c.
	// if err := notify.Watch("./...", c, notify.Remove); err != nil {
	for _, dir := range w.Directories {
		if err := notify.Watch(dir, c, w.ops...); err != nil {
			log.Printf("watch %s: %v", dir, err)
		}
	}
	defer notify.Stop(c)

	// Block until an event is received.
	for ei := range c {
		log.Printf("Got event: %v", ei)
		w.Do(ei.Path())
	}
}

func parseOperations(operations []string) []notify.Event {
	ops := make([]notify.Event, 0, len(operations))
	for _, op := range operations {
		switch strings.ToLower(op) {
		case "create":
			ops = append(ops, notify.Create)
		case "write":
			ops = append(ops, notify.Write)
		case "remove":
			ops = append(ops, notify.Remove)
		case "rename":
			ops = append(ops, notify.Rename)
		default:
			log.Printf("unknown operation %s", op)
		}
	}
	return ops
}
