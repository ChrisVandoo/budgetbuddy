package prompt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

type wizardStep int

const (
	stepName wizardStep = iota
	stepDateColumn
	stepDateFormat
	stepDescColumn
	stepAmountType
	stepSingleColumn
	stepSingleSign
	stepDualInColumn
	stepDualOutColumn
	stepDone
)

type SourceWizard struct {
	headers  []string
	filename string
	step     wizardStep
	inputs   []textinput.Model
	inputIdx int
	config   types.SourceConfig
	done     bool
	err      error
}

func NewSourceWizard(headers []string, filename string) *SourceWizard {
	makeInput := func(placeholder string) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.Focus()
		ti.CharLimit = 100
		ti.Width = 60
		return ti
	}

	return &SourceWizard{
		headers:  headers,
		filename: filename,
		step:     stepName,
		inputs:   []textinput.Model{makeInput("e.g., My Bank")},
		config: types.SourceConfig{
			Mapping: types.SourceMapping{},
		},
	}
}

func (m *SourceWizard) Init() tea.Cmd {
	return textinput.Blink
}

func (m *SourceWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case m.step == stepAmountType && msg.String() == "1":
			m.step = stepSingleColumn
			m.inputs = []textinput.Model{m.makeDropdown("Amount")}
			m.inputIdx = 0
			return m, nil
		case m.step == stepAmountType && msg.String() == "2":
			m.step = stepDualInColumn
			m.inputs = []textinput.Model{m.makeDropdown("Money In")}
			m.inputIdx = 0
			return m, nil
		case m.step == stepSingleSign && msg.String() == "y":
			m.config.Mapping.Amount.IsPositiveMoneyIn = true
			m.step = stepDone
			m.done = true
			return m, tea.Quit
		case m.step == stepSingleSign && msg.String() == "n":
			m.config.Mapping.Amount.IsPositiveMoneyIn = false
			m.step = stepDone
			m.done = true
			return m, tea.Quit
		case msg.String() == "ctrl+c", msg.String() == "q":
			m.done = true
			m.err = fmt.Errorf("cancelled")
			return m, tea.Quit
		case msg.String() == "enter":
			return m.handleEnter()
		case msg.String() == "up":
			return m.handleUp(), nil
		case msg.String() == "down":
			return m.handleDown(), nil
		}
	}

	if m.inputIdx < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.inputIdx], cmd = m.inputs[m.inputIdx].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *SourceWizard) handleUp() tea.Model {
	return m
}

func (m *SourceWizard) handleDown() tea.Model {
	return m
}

func (m *SourceWizard) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepName:
		m.config.Name = strings.TrimSpace(m.inputs[0].Value())
		if m.config.Name == "" {
			return m, nil
		}
		m.step = stepDateColumn
		m.inputs = []textinput.Model{m.makeDropdown("Transaction Date")}
		m.inputIdx = 0

	case stepDateColumn:
		col := strings.TrimSpace(m.inputs[0].Value())
		if col == "" {
			return m, nil
		}
		m.config.Mapping.Date.Header = col
		m.step = stepDateFormat
		m.inputs = []textinput.Model{m.makeDropdown("2006-01-02")}
		m.inputIdx = 0

	case stepDateFormat:
		fmtStr := strings.TrimSpace(m.inputs[0].Value())
		if fmtStr == "" {
			return m, nil
		}
		m.config.Mapping.Date.Format = fmtStr
		m.step = stepDescColumn
		m.inputs = []textinput.Model{m.makeDropdown("Description")}
		m.inputIdx = 0

	case stepDescColumn:
		col := strings.TrimSpace(m.inputs[0].Value())
		if col == "" {
			return m, nil
		}
		m.config.Mapping.Description.Header = col
		m.step = stepAmountType
		m.inputs = nil
		m.inputIdx = 0

	case stepAmountType:
		return m, nil

	case stepSingleColumn:
		col := strings.TrimSpace(m.inputs[0].Value())
		if col == "" {
			return m, nil
		}
		amt := &m.config.Mapping.Amount
		amt.SingleColumn = true
		amt.HeaderOut = col
		amt.HeaderIn = col
		m.step = stepSingleSign
		m.inputs = nil
		m.inputIdx = 0

	case stepSingleSign:
		return m, nil

	case stepDualInColumn:
		inCol := strings.TrimSpace(m.inputs[0].Value())
		if inCol == "" {
			return m, nil
		}
		m.config.Mapping.Amount.HeaderIn = inCol
		m.step = stepDualOutColumn
		m.inputs = []textinput.Model{m.makeDropdown("Funds Out")}
		m.inputIdx = 0

	case stepDualOutColumn:
		outCol := strings.TrimSpace(m.inputs[0].Value())
		if outCol == "" {
			return m, nil
		}
		m.config.Mapping.Amount.HeaderOut = outCol
		m.step = stepDone
		m.done = true
		return m, tea.Quit
	}

	return m, nil
}

func (m *SourceWizard) makeDropdown(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 60
	return ti
}

func (m *SourceWizard) View() string {
	if m.done {
		if m.err != nil {
			return "Source creation cancelled.\n"
		}
		return fmt.Sprintf("Source '%s' created successfully.\n", m.config.Name)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Create Bank Source for: %s\n", m.filename))
	b.WriteString(strings.Repeat("-", 40))
	b.WriteString("\n\n")

	switch m.step {
	case stepName:
		b.WriteString("Enter a name for this source:\n")
		b.WriteString(m.inputs[0].View())
		b.WriteString("\n\nHeaders detected:\n")
		for _, h := range m.headers {
			b.WriteString(fmt.Sprintf("  - %s\n", h))
		}

	// TODO: Should filter while typing based on listed columns to avoid typos
	case stepDateColumn:
		b.WriteString("Which column contains the transaction date?\n")
		b.WriteString(m.inputs[0].View())
		b.WriteString("\n\nAvailable columns:\n")
		for _, h := range m.headers {
			b.WriteString(fmt.Sprintf("  - %s\n", h))
		}

	// TODO: not clear what the format should be entered as, probably something like mm/dd/yyyy instead of actual values, should also show a row of data from the CSV
	case stepDateFormat:
		b.WriteString("Enter the date format (Go reference time: Mon Jan 2 15:04:05 MST 2006):\n")
		b.WriteString("Common formats:\n")
		b.WriteString("  2006-01-02  (ISO 8601)\n")
		b.WriteString("  01/02/2006  (US style)\n")
		b.WriteString("  20060102    (compact)\n")
		b.WriteString(m.inputs[0].View())

	// TODO: also should filter based on available columns
	case stepDescColumn:
		b.WriteString("Which column contains the description?\n")
		b.WriteString(m.inputs[0].View())
		b.WriteString("\n\nAvailable columns:\n")
		for _, h := range m.headers {
			b.WriteString(fmt.Sprintf("  - %s\n", h))
		}

	case stepAmountType:
		b.WriteString("How is the amount represented?\n")
		b.WriteString("  1) Single column (one column with the amount)\n")
		b.WriteString("  2) Dual column (separate 'money in' and 'money out' columns)\n")
		b.WriteString("\nPress 1 or 2:\n")

	case stepSingleColumn:
		b.WriteString("Which column contains the amount?\n")
		b.WriteString(m.inputs[0].View())
		b.WriteString("\n\nAvailable columns:\n")
		for _, h := range m.headers {
			b.WriteString(fmt.Sprintf("  - %s\n", h))
		}

	case stepSingleSign:
		b.WriteString("Does this column represent money coming in (positive)?\n")
		b.WriteString("  y) Yes - positive values are money in, negative are money out\n")
		b.WriteString("  n) No - positive values are money out, negative are money in\n")
		b.WriteString("\nPress y or n:\n")

	case stepDualInColumn:
		b.WriteString("Which column is 'money in'?\n")
		b.WriteString(m.inputs[0].View())

	case stepDualOutColumn:
		b.WriteString("Which column is 'money out'?\n")
		b.WriteString(m.inputs[0].View())
	}

	return b.String()
}

func (m *SourceWizard) Config() types.SourceConfig {
	return m.config
}

func (m *SourceWizard) Cancelled() bool {
	return m.err != nil
}
