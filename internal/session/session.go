package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/winkeep/winkeep/internal/process"
)

type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Processes []string  `json:"process_ids"`
}

type Manager struct {
	baseDir    string
	procMgr    *process.Manager
}

func NewManager(baseDir string, procMgr *process.Manager) *Manager {
	return &Manager{
		baseDir: filepath.Join(baseDir, "sessions"),
		procMgr: procMgr,
	}
}

func (m *Manager) Create(name string) (*Session, error) {
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
	s := &Session{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
		Processes: []string{},
	}

	if err := m.save(s); err != nil {
		return nil, err
	}
	return s, nil
}

func (m *Manager) List() ([]*Session, error) {
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sessions []*Session
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		s, err := m.load(filepath.Join(m.baseDir, entry.Name()))
		if err != nil {
			continue
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (m *Manager) Get(id string) (*Session, error) {
	return m.load(filepath.Join(m.baseDir, id))
}

func (m *Manager) GetByName(name string) (*Session, error) {
	sessions, err := m.List()
	if err != nil {
		return nil, err
	}
	for _, s := range sessions {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, fmt.Errorf("session %q not found", name)
}

func (m *Manager) AddProcess(sessionID, processID string) error {
	s, err := m.Get(sessionID)
	if err != nil {
		return err
	}
	s.Processes = append(s.Processes, processID)
	return m.save(s)
}

func (m *Manager) Kill(id string) error {
	s, err := m.Get(id)
	if err != nil {
		return err
	}

	for _, pid := range s.Processes {
		m.procMgr.Kill(pid)
	}

	return os.RemoveAll(filepath.Join(m.baseDir, id))
}

func (m *Manager) KillByName(name string) error {
	s, err := m.GetByName(name)
	if err != nil {
		return err
	}
	return m.Kill(s.ID)
}

func (m *Manager) save(s *Session) error {
	dir := filepath.Join(m.baseDir, s.ID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "meta.json"), data, 0644)
}

func (m *Manager) load(dir string) (*Session, error) {
	data, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
