package store

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestLedgerRollbackRestartAndExclusiveLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	memory, closeStore, err := NewFileMemory(path, InitialState())
	if err != nil {
		t.Fatal(err)
	}
	if _, closeOther, err := NewFileMemory(path, InitialState()); err == nil {
		closeOther()
		t.Fatal("concurrent process acquired ledger")
	}
	if err := memory.Update(func(state *State) error { state.CLIP.Balance = 99; return errors.New("abort") }); err == nil {
		t.Fatal("expected rollback")
	}
	if memory.Snapshot().CLIP.Balance != 0 {
		t.Fatal("partial write escaped rollback")
	}
	if err := memory.Update(func(state *State) error { state.CLIP.Balance = 12; return nil }); err != nil {
		t.Fatal(err)
	}
	closeStore()
	restarted, closeRestart, err := NewFileMemory(path, InitialState())
	if err != nil {
		t.Fatal(err)
	}
	defer closeRestart()
	if restarted.Snapshot().CLIP.Balance != 12 {
		t.Fatal("ledger lost after restart")
	}
}
