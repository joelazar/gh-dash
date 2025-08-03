package ui

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/cli/go-gh/v2/pkg/browser"

	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

func (m *Model) openBrowser() tea.Cmd {
	taskId := fmt.Sprintf("open_browser_%d", time.Now().Unix())
	log.Debug("openBrowser: starting browser open task", "taskId", taskId)
	task := context.Task{
		Id:           taskId,
		StartText:    "Opening in browser",
		FinishedText: "Opened in browser",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.ctx.StartTask(task)
	openCmd := func() tea.Msg {
		log.Debug("openBrowser: creating browser instance")
		b := browser.New("", os.Stdout, os.Stdin)
		currRow := m.getCurrRowData()
		log.Debug("openBrowser: got current row data", "currRow", currRow)
		if currRow == nil || reflect.ValueOf(currRow).IsNil() {
			log.Debug("openBrowser: current selection has no URL")
			return constants.TaskFinishedMsg{TaskId: taskId, Err: errors.New("current selection doesn't have a URL")}
		}
		log.Debug("openBrowser: current selection has URL")
		url := currRow.GetUrl()
		log.Debug("openBrowser: got URL", "url", url)
		log.Debug("openBrowser: attempting to browse URL", "url", url)
		err := b.Browse(url)
		if err != nil {
			log.Debug("openBrowser: error browsing URL", "error", err)
		} else {
			log.Debug("openBrowser: successfully opened URL in browser")
		}
		return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
	}
	return tea.Batch(startCmd, openCmd)
}
