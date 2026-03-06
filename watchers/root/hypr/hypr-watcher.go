package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"path"
	"time"

	"github.com/Iyed-M/watcher/shared"
	"github.com/shirou/gopsutil/process"
	"github.com/thiagokokada/hyprland-go"
	"github.com/thiagokokada/hyprland-go/event"
)

// TODO: move this ./linux_dial.go, make windows_dial.go and toggle between them with go:build linux
func DialSocket() (net.Conn, error) {
	return net.Dial("unix", shared.UnixSocketPath)
}

func Run() error {
	daemonConn, err := DialSocket()
	if err != nil {
		log.Fatalf("error connecting to the daemon make sure that the daemon is running. error: %v", err)
	}

	tick := flag.Duration("tick", 5*time.Second, "tick duation in seconds")
	flag.Parse()

	hyprClient := hyprland.MustClient()
	eventsClient := event.MustClient()
	ticker := time.NewTicker(*tick)
	for {
		var event shared.WatcherEvent
		ctx := context.Background()
		events, err := eventsClient.Receive(ctx)
		var activeWindows []string
		if err != nil {
			return err
		}

	}

	for {

		<-ticker.C
		window, err := hyprClient.ActiveWindow()
		if err != nil {
			return err
		}
		if window.Pid == 0 {
			log.Fatalf("got PID=0 window:%+v\n", window)
		}
		if err := hanldeWindow(window); err != nil {
			return err
		}
		slog.Info("ActiveWindow()", "pid", window.Pid, "title", window.Title, "class", window.Class, "address", window.Address)
	}
}

func sendEvent(conn net.Conn, event shared.WatcherEvent) error {
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(&event); err != nil {
		return err
	}

	return nil
}

func hanldeWindow(win hyprland.Window) error {
	var event shared.WatcherEvent
	ptree, err := inspectProcess(int32(win.Pid))
	if err != nil {
		slog.Error("can't inspectProcess", "err", err)
		ptree.Fprint(os.Stderr)
		return err
	}
	ptree.Fprint(os.Stdout)
	return nil
}

type processTree struct {
	Cmd      string
	CWD      string
	Children []processTree
	PID      int32
}

func (pt processTree) Fprint(w io.Writer) {
	pt.fprintRecursive(w, "", true, true)
}

func (pt processTree) fprintRecursive(w io.Writer, prefix string, isLast bool, isRoot bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	if isRoot {
		connector = ""
	}

	fmt.Fprintf(w, "%s%s%s \033[90m(%s)[%d]\033[0m\n", prefix, connector, pt.Cmd, pt.CWD, pt.PID)

	newPrefix := prefix
	if !isRoot {
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
	}

	for i, child := range pt.Children {
		isChildLast := i == len(pt.Children)-1
		child.fprintRecursive(w, newPrefix, isChildLast, false)
	}
}

func inspectProcess(pid int32) (*processTree, error) {
	ptree := &processTree{PID: pid}
	p, err := process.NewProcess(pid)
	if err != nil {
		return ptree, err
	}

	cwd, err := p.Cwd()
	if err != nil {
		ptree.CWD = "ERROR"
		return ptree, err
	}
	ptree.CWD = cwd

	cmd, err := p.Exe()
	if err != nil {
		ptree.Cmd = "ERROR"
		return ptree, err
	}
	ptree.Cmd = path.Base(cmd)

	var inspectedChildren []processTree
	children, err := p.Children()
	if err != nil && !errors.Is(err, process.ErrorNoChildren) {
		slog.Error("error reading children", "pid", pid)
		return ptree, err
	}
	for _, child := range children {
		inspected, err := inspectProcess(child.Pid)
		if err != nil {
			return inspected, err
		}
		inspectedChildren = append(inspectedChildren, *inspected)
	}

	ptree.Children = inspectedChildren
	return ptree, nil
}

func main() {
	log.Fatal(Run())
}
