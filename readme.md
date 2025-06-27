# ⏱️ Tickletimer

**Tickletimer** is a terminal-based multi-stopwatch application built with [Bubbletea](https://github.com/charmbracelet/bubbletea) — designed for speed, clarity, and a little flair. Track multiple timers with animations, key-driven control, and persistent storage across sessions.

![screenshot](https://user-images.githubusercontent.com/your-screenshot-here.png)

---

## ✨ Features

- ⏲️ Track unlimited named timers
- ⬆️⬇️ Navigate between timers
- 🎮 Keyboard-only interface
- 💾 Persistent storage in `~/.config/tickletimer/`
- 📁 CSV exports with timestamped dumps
- 🎨 Theme-aware styling using ANSI-safe colors
- 🔧 Editable timers: rename, reset, adjust time
- ⚡ Minimal system resource usage (I think. I didn't benchmark) 

---

## 🛠 Installation

```bash
git clone https://github.com/NicholasDewberryOfficial/TickleTimer
cd tickletimer
go build -o tickletimer
./tickletimer
```
⌨️ Keybindings

| Key     | Action                                |
| ------- | ------------------------------------- |
| `↑`/`↓` | Move between timers                   |
| `s`     | Start/stop selected timer             |
| `r`     | Enter Remove/Reset mode               |
| `a`     | Enter Add/Edit mode                   |
| `u` `u` | Dump all timers to CSV + prompt reset |
| `q`     | Quit and save                         |

🟥 Remove/Reset Mode
| Key       | Action                      |
| --------- | --------------------------- |
| `d`       | Delete timer (requires `y`) |
| `t`       | Reset timer (requires `y`)  |
| `[` / `]` | Adjust timer by ±30 seconds |
| `r`       | Return to normal mode       |

🟩 Add/Edit Mode
| Key       | Action                      |
| --------- | --------------------------- |
| `a`       | Add a new timer             |
| `r`       | Rename selected timer       |
| `[` / `]` | Adjust timer by ±30 seconds |
| `b`       | Return to normal mode       |

📂 Configuration

~/.config/tickletimer/config.json

Default config (if you don't have this line, you should paste it in):
``
{
  "enable_animations": true
}
``

Btw i only have 1 setting. Would yall like any more? Just ask.

💬 Acknowledgments

Built with:

    Bubbletea

    Lipgloss

    Bubbles

    The libre software community

Inspired by the desire to track time without distraction.

