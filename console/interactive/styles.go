package interactive

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var styleImpl = createStyles()

// =============================== styles =============================== //

type styles struct {
	titleStyle        lipgloss.Style
	itemStyle         lipgloss.Style
	focusedStyle      lipgloss.Style
	selectedItemStyle lipgloss.Style
	paginationStyle   lipgloss.Style
	helpStyle         lipgloss.Style
	quitTextStyle     lipgloss.Style
}

func createStyles() styles {
	return styles{
		titleStyle:        lipgloss.NewStyle().MarginLeft(2),
		itemStyle:         lipgloss.NewStyle().PaddingLeft(4),
		focusedStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
		selectedItemStyle: lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")),
		paginationStyle:   list.DefaultStyles().PaginationStyle.PaddingLeft(4),
		helpStyle:         list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1),
		quitTextStyle:     lipgloss.NewStyle().Margin(1, 0, 2, 2),
	}
}

// =============================== listItem =============================== //

type listItem string

func (o listItem) FilterValue() string { return "" }

// =============================== ListDelegate =============================== //

type listDelegate struct {
	styles styles
}

func (l listDelegate) Height() int { return 1 }

func (l listDelegate) Spacing() int { return 0 }

func (l listDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (l listDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if i, ok := item.(listItem); ok {
		str := fmt.Sprintf("%d. %s", index+1, i)

		fn := l.styles.itemStyle.Render
		if index == m.Index() {
			fn = func(s ...string) string {
				return l.styles.selectedItemStyle.Render("> " + strings.Join(s, " "))
			}
		}

		fmt.Fprint(w, fn(str))
	}
}
