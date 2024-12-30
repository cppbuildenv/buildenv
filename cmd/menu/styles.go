package menu

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#1976d2")).Padding(0, 1).Bold(true)
	inputStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF06B7"))
	actionBarStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	focusedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#9c27b0"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

// =============================== listItem =============================== //

type listItem string

func (o listItem) FilterValue() string { return "" }

// =============================== ListDelegate =============================== //

type listDelegate struct {
}

func (l listDelegate) Height() int { return 1 }

func (l listDelegate) Spacing() int { return 0 }

func (l listDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (l listDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if i, ok := item.(listItem); ok {
		str := fmt.Sprintf("%d. %s", index+1, i)

		fn := itemStyle.Render
		if index == m.Index() {
			fn = func(s ...string) string {
				return selectedItemStyle.Render("> " + strings.Join(s, " "))
			}
		}

		fmt.Fprint(w, fn(str))
	}
}
