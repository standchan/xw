package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type List struct {
	tviewList *tview.List
}

type Tui struct {
	titlePanel *tview.TextView
	listPanel  *List
	funcPanel  *tview.Flex
	mainPanel  *tview.Flex
	app        *tview.Application
}

func newTui() {
	ui := Tui{}
	ui.app = tview.NewApplication().EnableMouse(true)
	ui.titlePanel = ui.createTitlePanel()
	ui.listPanel = ui.createListPanel()
}

func (ui *Tui) createTitlePanel() *tview.TextView {
	return nil
}

func (ui *Tui) createListPanel() *List {
	return nil
}

func (ui *Tui) MouseCapture(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
	if event != nil || action == tview.MouseLeftClick {
		ui.listPanel.tviewList.GetCurrentItem()
		ui.listPanel.tviewList.SetSelectedFunc()
	}
}

func (ui *Tui) createFuncPanel() *tview.Flex {
	return nil
}

func (ui *Tui) createMainPanel() *tview.Flex {
	return nil
}
