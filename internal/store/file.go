package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// NewFileMemory is a durable single-process ledger. The lock prevents two
// servers from independently settling the same account or external request.
func NewFileMemory(path string, initial State) (*Memory, func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, nil, err
	}
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, nil, err
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lock.Close()
		return nil, nil, fmt.Errorf("ledger is already in use: %w", err)
	}
	closeStore := func() { _ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN); _ = lock.Close() }
	data, err := os.ReadFile(path)
	if err == nil {
		var saved State
		if err = json.Unmarshal(data, &saved); err != nil {
			closeStore()
			return nil, nil, fmt.Errorf("invalid ledger: %w", err)
		}
		if saved.CLIP.Symbol != "CLIP" || saved.Session.UserRef == "" || saved.Assets == nil || saved.Bindings == nil || saved.ExternalMigrations == nil {
			closeStore()
			return nil, nil, fmt.Errorf("incomplete ledger; refusing to replace existing records")
		}
		initial = saved
	} else if !os.IsNotExist(err) {
		closeStore()
		return nil, nil, err
	}
	memory := NewMemory(initial)
	memory.state.Session.Persistence = "local-file"
	for i := range memory.state.MigrationRequests {
		if memory.state.MigrationRequests[i].Status == "processing" {
			memory.state.MigrationRequests[i].Status = "attention_required"
			memory.state.MigrationRequests[i].Reason = "Interrupted execution; retry the original request to reconcile"
		}
	}
	memory.path = path
	if err := persistState(path, memory.state); err != nil {
		closeStore()
		return nil, nil, err
	}
	return memory, closeStore, nil
}

func persistState(path string, state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".clipli-ledger-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
