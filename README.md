<div align="center">

# ⚡ WinKeep

### nohup for Windows

**Keep your processes running after SSH disconnects.**

[![GitHub release](https://img.shields.io/github/v/release/kelvinzer0/winkeep?style=flat-square&color=blue)](https://github.com/kelvinzer0/winkeep/releases)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/kelvinzer0/winkeep?style=flat-square)](https://github.com/kelvinzer0/winkeep/stargazers)
[![GitHub issues](https://img.shields.io/github/issues/kelvinzer0/winkeep?style=flat-square)](https://github.com/kelvinzer0/winkeep/issues)
[![Downloads](https://img.shields.io/github/downloads/kelvinzer0/winkeep/total?style=flat-square&color=orange)](https://github.com/kelvinzer0/winkeep/releases)

[Install](#install) • [Usage](#usage) • [Why WinKeep?](#why-winkeep) • [Contributing](#contributing)

</div>

---

## The Problem

You SSH into your Windows server, start a long-running process, close your laptop, and... **everything dies.**

Windows has no `nohup`. No `screen`. No `tmux`. Your processes are tied to your SSH session.

## The Solution

```bash
winkeep run -- python server.py
```

**That's it.** Close SSH. Come back tomorrow. Your process is still running.

---

## Install

### PowerShell (Recommended)
```powershell
irm https://winkeep.dev/install.ps1 | iex
```

### Command Prompt
```cmd
curl -L https://winkeep.dev/install.bat -o install.bat && install.bat
```

### Go
```bash
go install github.com/kelvinzer0/winkeep/cmd/winkeep@latest
```

### Manual Download
Download the latest binary from [GitHub Releases](https://github.com/kelvinzer0/winkeep/releases) or [SourceForge](https://sourceforge.net/projects/winkeep/files/).

---

## Usage

### Run a background process
```bash
winkeep run -- python server.py
winkeep run --name myapp -- node app.js
winkeep run --session myproject -- npm start
```

### Manage processes
```bash
winkeep list                        # List all background processes
winkeep logs <id>                   # View process output
winkeep logs <id> --tail 100        # Last 100 lines
winkeep kill <id>                   # Stop a process
```

### Session management
```bash
winkeep session create myproject    # Create a session
winkeep session list                # List sessions
winkeep session kill myproject      # Kill all processes in session
```

---

## Why WinKeep?

| Feature | WinKeep | nohup (Linux) | NSSM | pm2 |
|---------|---------|---------------|------|-----|
| Zero config | ✅ | ✅ | ❌ | ❌ |
| Single binary | ✅ | ✅ | ❌ | ❌ |
| Session management | ✅ | ❌ | ❌ | ✅ |
| Log capture | ✅ | ✅ | ✅ | ✅ |
| Works on Windows | ✅ | ❌ | ✅ | ⚠️ |
| No runtime required | ✅ | ✅ | ✅ | ❌ |

---

## How It Works

WinKeep uses OS-level process detachment:

- **Windows**: `CREATE_NEW_PROCESS_GROUP` + `DETACHED_PROCESS` flags
- **Linux/macOS**: `setsid` to create a new session

Process metadata and logs are stored in:
- Windows: `%APPDATA%\winkeep\`
- Linux/macOS: `~/.config/winkeep/`

---

## Build from Source

```bash
# Requires Go 1.22+
make build          # Build for current platform
make build-all      # Build for all platforms
make release        # Build release binaries
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=kelvinzer0/winkeep&type=Date)](https://star-history.com/#kelvinzer0/winkeep&Date)

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

<div align="center">

**Made with ❤️ for the Windows SSH community**

</div>
