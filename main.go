package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Timer struct {
	Label         string        `json:"label"`
	StartTime     time.Time     `json:"-"`
	Elapsed       time.Duration `json:"elapsed"`
	Running       bool          `json:"running"`
	LastResetTime time.Time     `json:"-"`
}

type Config struct {
	EnableAnimations bool `json:"enable_animations"`
}

type mode int

const (
	normal mode = iota
	removeReset
	addEdit
	renameMode
	dumpConfirm
)

var spinnerFrames = []rune{'|', '/', '-', '\\'}

type model struct {
	Timers        []Timer
	Cursor        int
	Mode          mode
	lastKey       string
	lastTime      time.Time
	confirming    bool
	confirmAction string
	SpinnerIndex  int
	TickCount     int
	Config        Config
	TextInput     textinput.Model
	RenamingIndex int
	DumpPressed   int
}

func configFilePath() string {
	dirname, err := os.UserConfigDir()
	if err != nil {
		dirname = "."
	}
	return filepath.Join(dirname, "tickletimer", "timers.json")
}

func settingsFilePath() string {
	dirname, err := os.UserConfigDir()
	if err != nil {
		dirname = "."
	}
	return filepath.Join(dirname, "tickletimer", "config.json")
}

func dumpFilePath() string {
	dirname, err := os.UserConfigDir()
	if err != nil {
		dirname = "."
	}
	dumpDir := filepath.Join(dirname, "tickletimer", "dumps")
	os.MkdirAll(dumpDir, 0755)
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return filepath.Join(dumpDir, fmt.Sprintf("dump_%s.csv", timestamp))
}

func SaveTimers(timers []Timer) error {
	path := configFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(timers, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func DumpTimers(timers []Timer) error {
	file, err := os.Create(dumpFilePath())
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	w.Write([]string{"Label", "Elapsed"})
	for _, t := range timers {
		w.Write([]string{t.Label, t.Elapsed.String()})
	}
	return nil
}

func LoadTimers() []Timer {
	path := configFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return []Timer{
			{Label: "Timer A"},
			{Label: "Timer B"},
			{Label: "Timer C"},
		}
	}
	var timers []Timer
	if err := json.Unmarshal(data, &timers); err != nil {
		return []Timer{}
	}
	now := time.Now()
	for i := range timers {
		if timers[i].Running {
			timers[i].StartTime = now
		}
	}
	return timers
}

func LoadConfig() Config {
	path := settingsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{EnableAnimations: true}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{EnableAnimations: true}
	}
	return cfg
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "New timer name..."
	ti.Focus()
	ti.CharLimit = 64
	return model{
		Timers:    LoadTimers(),
		Config:    LoadConfig(),
		TextInput: ti,
	}
}

type tickMsg time.Time

func (m model) Init() tea.Cmd {
	return tea.Every(time.Second/10, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch m.Mode {
		case removeReset:
			switch key {
			case "y":
				if m.confirming {
					if m.Mode == dumpConfirm && m.confirmAction == "reset" {
						for i := range m.Timers {
							m.Timers[i].Elapsed = 0
							m.Timers[i].Running = false
						}
						m.Mode = normal
						m.confirming = false
						m.confirmAction = ""
						return m, nil
					}

					if m.confirmAction == "delete" && m.Cursor < len(m.Timers) {
						m.Timers = append(m.Timers[:m.Cursor], m.Timers[m.Cursor+1:]...)
						if m.Cursor >= len(m.Timers) && m.Cursor > 0 {
							m.Cursor--
						}
					} else if m.confirmAction == "reset" && m.Cursor < len(m.Timers) {
						t := &m.Timers[m.Cursor]
						t.Elapsed = 0
						t.Running = false
						t.LastResetTime = time.Now()
					}
					m.confirming = false
					m.Mode = normal
				}

			case "n":
				if m.confirming {
					m.confirming = false
					m.Mode = normal
				}
			case "d":
				if !m.confirming {
					m.confirming = true
					m.confirmAction = "delete"
				}
			case "t":
				if !m.confirming {
					m.confirming = true
					m.confirmAction = "reset"
				}
			case "[":
				if m.Cursor < len(m.Timers) {
					m.Timers[m.Cursor].Elapsed -= 30 * time.Second
					if m.Timers[m.Cursor].Elapsed < 0 {
						m.Timers[m.Cursor].Elapsed = 0
					}
				}
			case "]":
				if m.Cursor < len(m.Timers) {
					m.Timers[m.Cursor].Elapsed += 30 * time.Second
				}
			case "r":
				if !m.confirming {
					m.Mode = normal
				}
			}
			return m, nil

		case dumpConfirm:
			switch key {
			case "y":
				if m.confirming && m.confirmAction == "reset" {
					for i := range m.Timers {
						m.Timers[i].Elapsed = 0
						m.Timers[i].Running = false
					}
					m.confirming = false
					m.Mode = normal
					m.confirmAction = ""
				}
			case "n":
				if m.confirming {
					m.confirming = false
					m.Mode = normal
					m.confirmAction = ""
				}
			}
			return m, nil

		case normal:
			switch key {
			case "ctrl+c", "q":
				for i := range m.Timers {
					if m.Timers[i].Running {
						m.Timers[i].Elapsed += time.Since(m.Timers[i].StartTime)
						m.Timers[i].Running = false
					}
				}
				SaveTimers(m.Timers)
				return m, tea.Quit
			case "up":
				if m.Cursor > 0 {
					m.Cursor--
				}
			case "down":
				if m.Cursor < len(m.Timers)-1 {
					m.Cursor++
				}
			case "s":
				if m.Cursor < len(m.Timers) {
					t := &m.Timers[m.Cursor]
					if t.Running {
						t.Elapsed += time.Since(t.StartTime)
					} else {
						t.StartTime = time.Now()
					}
					t.Running = !t.Running
				}
			case "r":
				m.Mode = removeReset
			case "u":
				if time.Since(m.lastTime) < time.Second && m.lastKey == "u" {
					DumpTimers(m.Timers) // dump to CSV
					m.Mode = dumpConfirm // enter dump confirm mode
					m.confirming = true
					m.confirmAction = "reset" // reuse existing confirmation logic
				} else {
					m.lastKey = "u"
					m.lastTime = time.Now()
				}
			case "a":
				m.Mode = addEdit
			}

		case addEdit:
			switch key {
			case "a":
				m.Timers = append(m.Timers, Timer{Label: fmt.Sprintf("Timer %d", len(m.Timers)+1)})
			case "r":
				m.Mode = renameMode
				m.RenamingIndex = m.Cursor
				m.TextInput.SetValue(m.Timers[m.Cursor].Label)
				m.TextInput.Focus()
			case "[":
				m.Timers[m.Cursor].Elapsed -= 30 * time.Second
				if m.Timers[m.Cursor].Elapsed < 0 {
					m.Timers[m.Cursor].Elapsed = 0
				}
			case "]":
				m.Timers[m.Cursor].Elapsed += 30 * time.Second
			case "b":
				m.Mode = normal
			}

		case renameMode:
			switch key {
			case "enter":
				m.Timers[m.RenamingIndex].Label = m.TextInput.Value()
				m.Mode = addEdit
			case "esc":
				m.Mode = addEdit
			default:
				m.TextInput, _ = m.TextInput.Update(msg)
			}
		}

	case tickMsg:
		if m.Config.EnableAnimations {
			m.TickCount++
			m.SpinnerIndex = (m.SpinnerIndex + 1) % len(spinnerFrames)
		}
		return m, tea.Every(time.Second/10, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	for i, t := range m.Timers {
		elapsed := t.Elapsed
		if t.Running {
			elapsed += time.Since(t.StartTime)
		}

		prefix := "  "
		if t.Running && m.Config.EnableAnimations {
			prefix = fmt.Sprintf("%c ", spinnerFrames[m.SpinnerIndex])
		}

		mins := int(elapsed.Minutes())
		secs := int(elapsed.Seconds()) % 60
		label := fmt.Sprintf("%02d:%02d %s", mins, secs, t.Label)

		style := lipgloss.NewStyle()

		// Mode-based background
		switch m.Mode {
		case removeReset:
			style = style.Background(lipgloss.Color("1")) // Red
		case addEdit:
			style = style.Background(lipgloss.Color("2")) // Green
		case dumpConfirm:
			style = style.Background(lipgloss.Color("3")) // Yellow
		}

		// Cursor highlight
		if i == m.Cursor {
			style = style.Reverse(true)
			if m.Mode != normal {
				style = style.Bold(true)
			}
		}

		// Animate boldness
		if t.Running && m.Config.EnableAnimations && m.TickCount%2 == 0 {
			style = style.Bold(true)
		}

		b.WriteString(style.Render(prefix + label))
		b.WriteString("\n")
	}

	switch m.Mode {
	case removeReset:
		if m.confirming {
			switch m.confirmAction {
			case "delete":
				b.WriteString("⚠️  Delete this timer? (y/n)\n")
			case "reset":
				b.WriteString("↺  Reset this timer? (y/n)\n")
			}
		} else {
			b.WriteString("[Remove/Reset Mode]   d: delete  t: reset  [: -30s  ]: +30s  r: back\n")
		}
	case addEdit:
		b.WriteString("[Add/Edit Mode] a: add  r: rename  [: -30s  ]: +30s  b: back\n")
	case renameMode:
		b.WriteString("Renaming: " + m.TextInput.View() + "\n")
	case dumpConfirm:
		if m.confirming {
			b.WriteString("📁 Timers dumped. Reset all timers? (y/n)\n")
		}
	default:
		b.WriteString("↑/↓: navigate  s: start/stop  a: add/edit  r: remove/reset  u: dump/reset  q: quit\n")
	}

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
	}
}
