package prompt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ChrisVandoo/budgetbuddy/internal/db"
	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

type CategorySelector struct {
	database      *db.DB
	transaction   types.Transaction
	textInput     textinput.Model
	categories    []types.Category
	filtered      []types.Category
	selected      int
	done          bool
	chosen        *int64
	useGlob       bool
	globPattern   string
	confirmGlob   bool
	confirmed     bool
	err           error
}

func NewCategorySelector(database *db.DB, txn types.Transaction) *CategorySelector {
	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 60

	cats, _ := database.ListCategories()

	return &CategorySelector{
		database:    database,
		transaction: txn,
		textInput:   ti,
		categories:  cats,
		filtered:    cats,
	}
}

func (m *CategorySelector) Init() tea.Cmd {
	return textinput.Blink
}

func (m *CategorySelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit
		case "enter":
			if m.confirmGlob {
				if m.selected == 0 {
					// Use glob
					m.globPattern = m.textInput.Value()
					if m.globPattern == "" {
						m.globPattern = "*" + m.transaction.Description + "*"
					}
				}
				// Create rule and assign category
				cat := m.filtered[m.selected]
				if m.confirmGlob && m.selected == 0 {
					m.database.CreateRule(cat.ID, m.globPattern)
				} else {
					m.database.CreateRule(cat.ID, m.transaction.Description)
				}
				m.chosen = &cat.ID
				m.done = true
				m.confirmed = true
				return m, tea.Quit
			}
			if len(m.filtered) > 0 {
				m.chosen = &m.filtered[m.selected].ID
				if m.textInput.Value() != "" {
					m.confirmGlob = true
					m.globPattern = m.textInput.Value()
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Enter glob pattern or press Enter for exact:"
					m.filtered = []types.Category{
						{Name: "Use current filter as glob pattern"},
						{Name: m.transaction.Description + " (exact match)"},
					}
					m.selected = 0
					return m, nil
				}
				// No filter text - use exact description as rule
				cat := m.filtered[m.selected]
				m.database.CreateRule(cat.ID, m.transaction.Description)
				m.chosen = &cat.ID
				m.done = true
				m.confirmed = true
				return m, tea.Quit
			}
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.filtered)-1 {
				m.selected++
			}
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	if !m.confirmGlob {
		m.filter()
	}

	return m, cmd
}

func (m *CategorySelector) filter() {
	filterText := strings.ToLower(strings.TrimSpace(m.textInput.Value()))
	if filterText == "" {
		m.filtered = m.categories
		return
	}

	var filtered []types.Category
	for _, cat := range m.categories {
		if strings.Contains(strings.ToLower(cat.Name), filterText) {
			filtered = append(filtered, cat)
		}
	}
	m.filtered = filtered
	if m.selected >= len(m.filtered) {
		m.selected = 0
	}
}

func (m *CategorySelector) View() string {
	if m.done {
		if m.confirmed {
			return fmt.Sprintf("Categorized as: %s\n", m.getCategoryName())
		}
		return "Cancelled.\n"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Transaction: %s ($%d.%02d)\n",
		m.transaction.Description,
		m.transaction.AmountCents/100,
		abs(m.transaction.AmountCents%100)))

	if m.confirmGlob {
		b.WriteString("\nCreate rule as glob pattern?\n")
		b.WriteString(fmt.Sprintf("Description: %s\n", m.transaction.Description))
		b.WriteString(fmt.Sprintf("Filter text: %s\n", m.globPattern))
		b.WriteString("Press Enter to confirm, or type a custom glob pattern:\n\n")
		b.WriteString(m.textInput.View())
		b.WriteString("\n\n")
		for i, opt := range m.filtered {
			cursor := " "
			if i == m.selected {
				cursor = ">"
			}
			b.WriteString(fmt.Sprintf("%s %s\n", cursor, opt.Name))
		}
		return b.String()
	}

	b.WriteString("\nSelect a category:\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	if len(m.filtered) == 0 {
		b.WriteString("No matching categories. Press q to cancel.\n")
	} else {
		for i, cat := range m.filtered {
			cursor := " "
			if i == m.selected {
				cursor = ">"
			}
			b.WriteString(fmt.Sprintf("%s %s ($%d/mo)\n",
				cursor, cat.Name, cat.MonthlyBudgetCents/100))
		}
	}

	return b.String()
}

func (m *CategorySelector) getCategoryName() string {
	if m.chosen != nil {
		for _, cat := range m.categories {
			if cat.ID == *m.chosen {
				return cat.Name
			}
		}
	}
	return ""
}

func (m *CategorySelector) Chosen() *int64 {
	return m.chosen
}

func (m *CategorySelector) Confirmed() bool {
	return m.confirmed
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
