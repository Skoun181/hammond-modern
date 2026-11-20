package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type portableRecoveryMarker struct {
	SafetyDirectory string `json:"safetyDirectory"`
}

func portableJournalPath(config string) string {
	return filepath.Join(config, "restore-safety", "in-progress.json")
}

func portableSyncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func portableStartJournal(config, rescue string) error {
	root := filepath.Join(config, "restore-safety")
	if filepath.Dir(rescue) != root || !strings.HasPrefix(filepath.Base(rescue), "pre-restore-") {
		return errors.New("Ungültiges Rettungsverzeichnis")
	}
	target := portableJournalPath(config)
	if _, err := os.Stat(target); err == nil {
		return errors.New("Eine unterbrochene Wiederherstellung ist noch nicht bereinigt")
	}
	payload, err := json.Marshal(portableRecoveryMarker{SafetyDirectory: filepath.Base(rescue)})
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(root, ".hm-journal-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err = tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err = tmp.Write(payload); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmp.Name(), target); err != nil {
		return err
	}
	return portableSyncDir(root)
}

func portableFinishJournal(config string) error {
	p := portableJournalPath(config)
	if err := os.Remove(p); err != nil {
		return err
	}
	return portableSyncDir(filepath.Dir(p))
}

func portableRecoverFiles(assets, rescue string) error {
	source := filepath.Join(rescue, "assets")
	saved, err := portableAssets(source)
	if err != nil {
		return err
	}
	// Only overwrite the application's known regular asset files.
	existing, err := os.ReadDir(assets)
	if err != nil {
		return err
	}
	for _, item := range existing {
		name := item.Name()
		path := filepath.Join(assets, name)
		if strings.HasPrefix(name, ".hm-stage-") || strings.HasPrefix(name, ".hm-rollback-") {
			if err = os.RemoveAll(path); err != nil {
				return err
			}
			continue
		}
		if !portableAssetName(name) || !item.Type().IsRegular() {
			return fmt.Errorf("Unerwarteter Dateityp während der Notfallwiederherstellung: %s", name)
		}
		if err = os.Remove(path); err != nil {
			return err
		}
	}
	for _, name := range saved {
		if err = portableCopy(filepath.Join(source, name), filepath.Join(assets, name)); err != nil {
			return err
		}
	}
	return nil
}

// RecoverPortableRestore runs BEFORE db.Init(), while no SQLite connection is
// open. It reverts even a crash between asset changes and SQLite commit.
func RecoverPortableRestore() error {
	config, assets, err := portablePaths()
	if err != nil {
		return err
	}
	markerPath := portableJournalPath(config)
	payload, err := os.ReadFile(markerPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var marker portableRecoveryMarker
	if err = json.Unmarshal(payload, &marker); err != nil {
		return err
	}
	name := marker.SafetyDirectory
	if !strings.HasPrefix(name, "pre-restore-") || name != filepath.Base(name) || strings.ContainsAny(name, "/\\") {
		return errors.New("Ungültiger Recovery-Journal-Eintrag")
	}
	rescue := filepath.Join(config, "restore-safety", name)
	snapshot := filepath.Join(rescue, "hammond.db")
	check, err := sql.Open("sqlite3", "file:"+snapshot+"?mode=ro")
	if err != nil {
		return err
	}
	if err = portableCheckSQLite(check); err != nil {
		check.Close()
		return err
	}
	check.Close()
	log.Printf("interrupted portable restore found: recovering previous data from %s", rescue)
	// Keep the journal until *both* database and files are restored. Recovery
	// is idempotent and can safely resume after another unplanned interruption.
	if err = portableRecoverFiles(assets, rescue); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(config, ".hm-db-recovery-")
	if err != nil {
		return err
	}
	tempPath := tmp.Name()
	defer os.Remove(tempPath)
	src, err := os.Open(snapshot)
	if err != nil {
		tmp.Close()
		return err
	}
	_, err = io.Copy(tmp, src)
	src.Close()
	if err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	// Old WAL/shm files belong to the interrupted live database, not the
	// recovered consistent SQLite snapshot.
	for _, suffix := range []string{"-wal", "-shm"} {
		if err = os.Remove(filepath.Join(config, "hammond.db") + suffix); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err = os.Rename(tempPath, filepath.Join(config, "hammond.db")); err != nil {
		return err
	}
	if err = portableSyncDir(config); err != nil {
		return err
	}
	if err = portableFinishJournal(config); err != nil {
		return err
	}
	log.Print("portable restore recovery finished")
	return nil
}
