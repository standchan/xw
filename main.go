package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/rivo/tview"
	"os"
	"os/exec"
	"strings"
)

const confDir string = "/usr/local/etc/xwguard/"

var (
	app        *tview.Application // The tview application.
	pages      *tview.Pages       // The application pages.
	finderPage = "*finder*"       // The name of the Finder page.
)

// TODO 这个系统类型的表达不是很好
// TODO 尝试使用错误码
// TODO 这里不应该是创建文件夹，应该是判断文件夹存在与否
var systemType uint16 = 0

type guardfile struct {
	rundir   string
	runcmd   string
	match    string
	profiles string
}

// eg: map[redis.guard]guardfile
type guardfileMap map[string]guardfile

type processInfo struct {
	name string
	pid  string
}

type userControl struct {
	cmdType     string
	processInfo processInfo
}

var processInfoChan = make(chan processInfo, 30)
var userRequestChan = make(chan userControl, 10)
var errChan = make(chan error, 0)

func main() {
	front()
	//egFlex()
	//egGrid()
	//egList()
}

func egList() {
	app := tview.NewApplication()
	list := tview.NewList().ShowSecondaryText(false).
		AddItem("List item 1", "", 0, nil).
		AddItem("List item 2", "", 0, nil).
		AddItem("List item 3", "", 0, nil).
		AddItem("List item 4", "", 0, nil).
		AddItem("Quit", "Press to exit", 0, func() {
			app.Stop()
		})
	if err := app.SetRoot(list, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}

func egFlex() {
	app := tview.NewApplication()
	flex := tview.NewFlex().
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Left (1/2 x width of Top)"), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Top"), 0, 1, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)"), 0, 3, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom (5 rows)"), 5, 1, false), 0, 2, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Right (20 cols)"), 20, 1, false)
	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}
}

func egGrid() {
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}
	menu := newPrimitive("Menu")
	main := newPrimitive("Main content")
	sideBar := newPrimitive("Side Bar")

	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0, 30).
		SetBorders(true).
		AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false).
		AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false).
		AddItem(main, 1, 0, 1, 3, 0, 0, false).
		AddItem(sideBar, 0, 0, 0, 0, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(menu, 1, 0, 1, 1, 0, 100, false).
		AddItem(main, 1, 1, 1, 1, 0, 100, false).
		AddItem(sideBar, 1, 2, 1, 1, 0, 100, false)

	if err := tview.NewApplication().SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}

func backend() {
	gmap, err := readGuardfile(confDir)
	if err != nil {
		panic(err)
	}
	for i, v := range gmap {
		var p processInfo
		p.name = i
		pid, err := getPid(v.match)
		if err != nil {
			p.pid = pid
		}
		processInfoChan <- p
	}
	go func() {
		select {
		case req := <-userRequestChan:
			handleUserRequest(req)
		}
	}()
}

func handleUserRequest(control userControl) {

}

func handleProcess(pid string, cmdType string) (err error) {
	if cmdType == "kill" {
		_, err := exec.Command("kill", pid).Output()
		if err != nil {
			return err
		}
	}
	if cmdType == "kill9" {
		_, err := exec.Command("kill", "-9", pid).Output()
		if err != nil {
			return err
		}
	}
	if cmdType == "restart" {
		_, err := exec.Command("kill", pid).Output()
		if err != nil {
			return err
		}
		//TODO 还需要重启
	}
	return err
}

func front() {
	newUI()
}

func mkdirConfDir() {
	err := os.MkdirAll(confDir, os.ModeDir)
	if err != nil {
		errChan <- err
	}
	os.Exit(1)
}

func getSystemType() {
	output, err := exec.Command("cat", "/etc/issue|grep", "-c", "Ubuntu").Output()
	if err != nil {
		errChan <- err
		os.Exit(1)
	}
	systemType = binary.BigEndian.Uint16(output)
}

// TODO
// footer按钮之间需要隔开，然后增加按钮事件。
func newUI() {
	app = tview.NewApplication().EnableMouse(true)
	header := tview.NewTextView().SetTextAlign(tview.AlignLeft).SetText("        guardfile          pids\n----------------------------------")
	header.SetBorder(true)

	newButton := func(name string) tview.Primitive {
		return tview.NewButton(name)
	}
	menu := tview.NewList().ShowSecondaryText(false).
		AddItem("redis      9000", "", 0, nil).
		AddItem("kafka      9230", "", 0, nil)
	menu.SetBorder(false)

	footer := tview.NewFlex().SetDirection(tview.FlexRowCSS).
		AddItem(newButton("F1HELP"), 8, 0, true).
		AddItem(newButton("F2KILL"), 8, 0, false).
		AddItem(newButton("F3RESTART"), 8, 0, false).
		AddItem(newButton("F4QUIT"), 8, 0, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 0, 2, false).
		AddItem(menu, 0, 2, true).
		AddItem(footer, 1, 1, false)
	app.SetRoot(flex, true)
	if err := app.Run(); err != nil {
		fmt.Printf("Error running application: %s\n", err)
	}
}

// 两种方案
// 1.执行linux命令获取相应进程信息
// 2.使用gopsutil库获取相应信息
func getGuardfile() chan<- map[string]string {
	return nil
}

func readGuardfile(dirPath string) (map[string]guardfile, error) {
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	var m = make(guardfileMap, 0)
	for _, v := range dir {
		if !v.IsDir() && strings.HasSuffix(v.Name(), ".guard") {
			//fs, err := os.ReadFile(fmt.Sprintf("%s%s", confDir, v.Name()))
			fs, err := os.Open(fmt.Sprintf("%s%s", confDir, v.Name()))
			if err != nil {
				errChan <- err
				continue
			}
			sc := bufio.NewScanner(fs)
			var g guardfile
			for sc.Scan() {
				//strings.HasPrefix()
				if strings.HasPrefix(sc.Text(), "rundir") {
					g.rundir = parseGuardfile(sc.Text())
				} else if strings.HasPrefix(sc.Text(), "runcmd") {
					g.runcmd = parseGuardfile(sc.Text())
				} else if strings.HasPrefix(sc.Text(), "match") {
					g.match = parseGuardfile(sc.Text())
				} else if strings.HasPrefix(sc.Text(), "profiles") {
					g.profiles = parseGuardfile(sc.Text())
				}
			}
			m[v.Name()] = g
		}
	}
	return m, nil
}

func parseGuardfile(s string) string {
	ss := strings.Split(s, "s")
	if len(ss) == 2 {
		return ss[1]
	} else {
		return ""
	}
}

func getPid(match string) (string, error) {
	output, err := exec.Command("ps", fmt.Sprintf("aux | grep %s | grep -v grep | awk -F' ' '{print $2}' | sed 's#\\n# #g'", match)).Output()
	if err != nil {
		return "", err
	}
	return string(output), err
}
