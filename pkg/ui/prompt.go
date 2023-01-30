package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Prompter struct {
	currentPos int
	inputs     []textinput.Model
	prompts    []*Prompt
}

func NewPrompter(prompts []*Prompt) *Prompter {
	textinputs := make([]textinput.Model, len(prompts))
	for i := range prompts {
		ti := textinput.New()
		ti.Prompt = prompts[i].text

		// Set default styles here etc
		if prompts[i].maskedValue {
			ti.EchoCharacter = '*'
			ti.EchoMode = textinput.EchoPassword
		}

		ti = applyDefaultLayout(ti)
		textinputs[i] = ti
	}

	textinputs[0].Focus()

	return &Prompter{
		inputs:     textinputs,
		prompts:    prompts,
		currentPos: 0,
	}
}

func applyDefaultLayout(ti textinput.Model) textinput.Model {
	promptStyle := lipgloss.NewStyle().Width(15)

	ti.PromptStyle = promptStyle
	ti.CharLimit = 32
	return ti
}

func (p *Prompter) Init() tea.Cmd {
	return textinput.Blink
}

type errMsg error

func (p *Prompter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return p, tea.Quit
		case tea.KeyEnter, tea.KeyTab:
			// If this is the last one, quit
			if p.currentPos == len(p.inputs)-1 {
				p.updateValues()
				return p, tea.Quit
			}

			p.nextInput()

			for i := range p.inputs {
				p.inputs[i].Blur()
			}
			p.inputs[p.currentPos].Focus()
		}

	// We handle errors just like any other message
	case errMsg:
		// p.err = msg
		return p, nil
	}

	// Handle character input and blinking
	return p, p.updateInputs(msg)
}

func (p *Prompter) updateValues() {
	for i := range p.inputs {
		p.prompts[i].rValue = p.inputs[i].Value()
	}
}

func (p *Prompter) nextInput() {
	p.currentPos = (p.currentPos + 1) % len(p.inputs)
}

func (p *Prompter) View() string {
	var sb strings.Builder
	for i := range p.inputs {
		if i > p.currentPos {
			break
		}
		sb.WriteString(p.inputs[i].View())
		if i < len(p.inputs) {
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}

func (p *Prompter) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(p.inputs))

	for i := range p.inputs {
		p.inputs[i], cmds[i] = p.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

type Prompt struct {
	text        string
	rValue      string
	maskedValue bool
}

func NewPrompt(text string) *Prompt {
	return &Prompt{
		text: text,
	}
}

func (p *Prompt) Value() string {
	return p.rValue
}

func (p *Prompt) Mask() *Prompt {
	p.maskedValue = true
	return p
}

// TODO Add ValidationFunc etc
