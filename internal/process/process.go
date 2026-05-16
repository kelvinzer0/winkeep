package process

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Status string

const (
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusExited  Status = "exited"
)

type Process struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	PID       int       `json:"pid"`
	Status    Status    `json:"status"`
	StartTime time.Time `json:"start_time"`
	ExitCode  int       `json:"exit_code,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	LogDir    string    `json:"log_dir"`
}

type Manager struct {
	baseDir string
}

func NewManager(baseDir string) *Manager {
	return &Manager{baseDir: filepath.Join(baseDir, "processes")}
}

func (m *Manager) Run(name, command string, args []string, sessionID string) (*Process, error) {
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create base dir: %w", err)
	}

	id := fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
	logDir := filepath.Join(m.baseDir, id)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	stdoutPath := filepath.Join(logDir, "stdout.log")
	stderrPath := filepath.Join(logDir, "stderr.log")

	stdoutFile, err := os.Create(stdoutPath)
	if err != nil {
		return nil, fmt.Errorf("create stdout log: %w", err)
	}

	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		stdoutFile.Close()
		return nil, fmt.Errorf("create stderr log: %w", err)
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	setDetachAttrs(cmd)

	if err := cmd.Start(); err != nil {
		stdoutFile.Close()
		stderrFile.Close()
		return nil, fmt.Errorf("start process: %w", err)
	}

	p := &Process{
		ID:        id,
		Name:      name,
		Command:   command,
		Args:      args,
		PID:       cmd.Process.Pid,
		Status:    StatusRunning,
		StartTime: time.Now(),
		SessionID: sessionID,
		LogDir:    logDir,
	}

	if err := m.saveMeta(p); err != nil {
		cmd.Process.Kill()
		return nil, err
	}

	go func() {
		err := cmd.Wait()
		stdoutFile.Close()
		stderrFile.Close()

		p.Status = StatusExited
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				p.ExitCode = exitErr.ExitCode()
			} else {
				p.ExitCode = -1
			}
		}
		m.saveMeta(p)
	}()

	return p, nil
}

func (m *Manager) List() ([]*Process, error) {
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var processes []*Process
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		p, err := m.loadMeta(filepath.Join(m.baseDir, entry.Name()))
		if err != nil {
			continue
		}
		processes = append(processes, p)
	}
	return processes, nil
}

func (m *Manager) Get(id string) (*Process, error) {
	return m.loadMeta(filepath.Join(m.baseDir, id))
}

func (m *Manager) Kill(id string) error {
	p, err := m.Get(id)
	if err != nil {
		return err
	}

	if p.Status != StatusRunning {
		return fmt.Errorf("process %s is not running", id)
	}

	killProcess(p.PID)

	p.Status = StatusStopped
	return m.saveMeta(p)
}

func (m *Manager) Logs(id string, tail int) (string, error) {
	p, err := m.Get(id)
	if err != nil {
		return "", err
	}

	stdoutPath := filepath.Join(p.LogDir, "stdout.log")
	data, err := os.ReadFile(stdoutPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	lines := splitLines(string(data))
	if tail > 0 && len(lines) > tail {
		lines = lines[len(lines)-tail:]
	}
	return joinLines(lines), nil
}

func (m *Manager) saveMeta(p *Process) error {
	metaPath := filepath.Join(m.baseDir, p.ID, "meta.json")
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, data, 0644)
}

func (m *Manager) loadMeta(dir string) (*Process, error) {
	metaPath := filepath.Join(dir, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var p Process
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}

	if p.Status == StatusRunning {
		if !isProcessAlive(p.PID) {
			p.Status = StatusExited
			m.saveMeta(&p)
		}
	}

	return &p, nil
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, l := range lines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}
