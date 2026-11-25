package service

// Portable backups contain a sanitised, consistent SQLite snapshot and the
// application assets. They deliberately never contain passwords or .env files.
import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/akhilrex/hammond/db"
	_ "github.com/mattn/go-sqlite3"
)

const (
	PortableBackupVersion       = 1
	PortableAppVersion          = "1.2.0"
	PortableMaxZIP        int64 = 512 << 20
	PortableMaxData       int64 = 1024 << 20
	PortableMaxFiles            = 10000
)

var PortableDataLock sync.RWMutex

const portableFormat = "hammond-modern-portable"

var safeTable = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
var assetExtensions = map[string]bool{
	".pdf": true, ".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	".gif": true, ".bmp": true, ".tif": true, ".tiff": true, ".heic": true,
	".txt": true, ".csv": true, ".doc": true, ".docx": true,
	".xls": true, ".xlsx": true,
}
var allowedTables = map[string]bool{
	"users": true, "vehicles": true, "user_vehicles": true,
	"fillups": true, "expenses": true, "attachments": true,
	"vehicle_attachments": true, "drivers": true, "maintenance_tasks": true,
	"tires": true, "tire_mileage_states": true, "maintenance_rules": true, "recurring_costs": true, "financing_terms": true, "financing_calculations": true, "home_assistant_keys": true, "settings": true, "quick_entries": true, "job_locks": true, "migrations": true,
	"vehicle_alerts": true, "alert_occurances": true, "notifications": true,
}

type PortableEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}
type PortableManifest struct {
	Format        string           `json:"format"`
	FormatVersion int              `json:"formatVersion"`
	AppVersion    string           `json:"appVersion"`
	SchemaVersion int              `json:"schemaVersion"`
	SchemaSHA256  string           `json:"schemaSHA256"`
	CreatedAt     time.Time        `json:"createdAt"`
	Files         []PortableEntry  `json:"files"`
	Counts        map[string]int64 `json:"counts"`
}
type PortableSummary struct {
	Manifest          PortableManifest `json:"manifest"`
	AssetCount        int              `json:"assetCount"`
	AdminEmailMatched bool             `json:"adminEmailMatched"`
	PasswordNotice    string           `json:"passwordNotice"`
	SafetyBackup      string           `json:"safetyBackup,omitempty"`
}

func portablePaths() (string, string, error) {
	config := os.Getenv("CONFIG")
	assets := os.Getenv("DATA")
	if config == "" || assets == "" {
		return "", "", errors.New("CONFIG oder DATA ist nicht gesetzt")
	}
	return config, assets, nil
}

func portableSnapshot(path string) error {
	// VACUUM INTO is a SQLite-native consistent snapshot, including WAL changes.
	if _, err := os.Stat(path); err == nil {
		return errors.New("Snapshot-Ziel existiert bereits")
	}
	sqldb, err := db.DB.DB()
	if err != nil {
		return err
	}
	_, err = sqldb.Exec("VACUUM INTO ?", path)
	if err != nil {
		return fmt.Errorf("SQLite Snapshot: %w", err)
	}
	return os.Chmod(path, 0600)
}

func portableCheckSQLite(conn *sql.DB) error {
	var result string
	if err := conn.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return errors.New("SQLite-Integritätsprüfung fehlgeschlagen")
	}
	return nil
}

func portableSchema(conn *sql.DB) (string, []string, error) {
	return portableSchemaOmitting(conn, nil)
}

// Only explicitly supported additive tables may be omitted by the caller.
func portableSchemaOmitting(conn *sql.DB, omit map[string]bool) (string, []string, error) {
	// No source-controlled views or triggers are needed for this app. Reject
	// unexpected SQL objects in untrusted backup databases.
	var unsafeObjects int
	if err := conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type IN ('view', 'trigger')").Scan(&unsafeObjects); err != nil {
		return "", nil, err
	}
	if unsafeObjects != 0 {
		return "", nil, errors.New("Die Datenbank enthält unerwartete SQL-Objekte")
	}
	rows, err := conn.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return "", nil, err
	}
	names := []string{}
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			break
		}
		if !safeTable.MatchString(table) || !allowedTables[table] {
			err = fmt.Errorf("Nicht unterstützte Datenbanktabelle: %s", table)
			break
		}
		if omit[table] { continue }
		names = append(names, table)
	}
	if closeErr := rows.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", nil, err
	}
	if len(names) == 0 {
		return "", nil, errors.New("Datenbank enthält keine Tabellen")
	}
	var b strings.Builder
	for _, table := range names {
		b.WriteString(table)
		b.WriteByte('\n')
		info, err := conn.Query(`PRAGMA table_info("` + table + `")`)
		if err != nil {
			return "", nil, err
		}
		columnCount := 0
		for info.Next() {
			var cid, notNull, pk int
			var name, typ string
			var def sql.NullString
			if err = info.Scan(&cid, &name, &typ, &notNull, &def, &pk); err != nil {
				break
			}
			b.WriteString(fmt.Sprintf("%d:%s:%s:%d:%d:%s\n", cid, name, typ, notNull, pk, def.String))
			columnCount++
		}
		if closeErr := info.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return "", nil, err
		}
		if columnCount == 0 {
			return "", nil, fmt.Errorf("Ungültige Tabelle %s", table)
		}
	}
	hash := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(hash[:]), names, nil
}

func portableSanitise(path string) (string, []string, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return "", nil, err
	}
	defer conn.Close()
	conn.SetMaxOpenConns(1)
	// Overwriting/deleting credential rows must also erase old SQLite page
	// content, not merely make the sensitive rows unreachable to SQL.
	if _, err = conn.Exec("PRAGMA secure_delete=ON"); err != nil {
		return "", nil, err
	}
	if err = portableCheckSQLite(conn); err != nil {
		return "", nil, err
	}
	schema, tables, err := portableSchema(conn)
	if err != nil {
		return "", nil, err
	}
	for _, table := range tables {
		// API keys are never portable: copying them would grant a different
		// server the same authorization. Do not even export their hashes.
		if table == "home_assistant_keys" {
			if _, err = conn.Exec("DELETE FROM home_assistant_keys"); err != nil {
				return "", nil, err
			}
			continue
		}
		rows, queryErr := conn.Query(`PRAGMA table_info("` + table + `")`)
		if queryErr != nil {
			return "", nil, queryErr
		}
		cols := []string{}
		for rows.Next() {
			var cid, notNull, pk int
			var name, typ string
			var def sql.NullString
			if queryErr = rows.Scan(&cid, &name, &typ, &notNull, &def, &pk); queryErr != nil {
				break
			}
			lower := strings.ToLower(name)
			if strings.Contains(lower, "password") || strings.Contains(lower, "secret") ||
				strings.Contains(lower, "token") || strings.Contains(lower, "api_key") ||
				strings.Contains(lower, "apikey") || strings.Contains(lower, "credential") {
				if !safeTable.MatchString(lower) {
					queryErr = errors.New("Unsichere Spaltenbezeichnung")
					break
				}
				cols = append(cols, name)
			}
		}
		if e := rows.Close(); queryErr == nil {
			queryErr = e
		}
		if queryErr != nil {
			return "", nil, queryErr
		}
		for _, col := range cols {
			if _, err = conn.Exec(`UPDATE "` + table + `" SET "` + col + `" = ''`); err != nil {
				return "", nil, fmt.Errorf("Zugangsdaten konnten nicht entfernt werden: %w", err)
			}
		}
	}
	// Rebuild the snapshot after all redactions. This also prevents credential
	// fragments in free pages from being included in the ZIP file.
	if _, err = conn.Exec("VACUUM"); err != nil {
		return "", nil, err
	}
	if err = portableCheckSQLite(conn); err != nil {
		return "", nil, err
	}
	return schema, tables, nil
}

func portableCount(conn *sql.DB, tables []string) (map[string]int64, error) {
	out := make(map[string]int64, len(tables))
	for _, table := range tables {
		var n int64
		if err := conn.QueryRow(`SELECT COUNT(*) FROM "` + table + `"`).Scan(&n); err != nil {
			return nil, err
		}
		out[table] = n
	}
	return out, nil
}
func portableAssetName(name string) bool {
	return name != "" && name != "." && name != ".." && name == filepath.Base(name) &&
		!strings.HasPrefix(name, ".") && !strings.ContainsAny(name, `/`) &&
		len(name) < 256 && assetExtensions[strings.ToLower(filepath.Ext(name))]
}
func portableAssets(assets string) ([]string, error) {
	files, err := os.ReadDir(assets)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, f := range files {
		if f.Name() == "@eaDir" || f.Name() == "#recycle" || strings.HasPrefix(f.Name(), ".hm-stage-") || strings.HasPrefix(f.Name(), ".hm-rollback-") {
			continue
		}
		if !portableAssetName(f.Name()) || !f.Type().IsRegular() {
			return nil, fmt.Errorf("Nicht unterstützte Datei unter /assets: %s", f.Name())
		}
		result = append(result, f.Name())
		if len(result) > PortableMaxFiles {
			return nil, errors.New("Zu viele Dateien im /assets-Verzeichnis")
		}
	}
	sort.Strings(result)
	return result, nil
}
func portableCheckReferences(conn *sql.DB, files map[string]bool) error {
	rows, err := conn.Query("SELECT path FROM attachments WHERE deleted_at IS NULL")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var p string
		if err = rows.Scan(&p); err != nil {
			return err
		}
		name := filepath.Base(p)
		if !portableAssetName(name) || !files[name] {
			return fmt.Errorf("Anhang fehlt oder Dateityp ist nicht unterstützt: %s", name)
		}
	}
	return rows.Err()
}
func portableZipFile(zw *zip.Writer, archiveName, sourcePath string) (PortableEntry, error) {
	var item PortableEntry
	file, err := os.Open(sourcePath)
	if err != nil {
		return item, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return item, err
	}
	if !stat.Mode().IsRegular() {
		return item, errors.New("Keine reguläre Datei: " + archiveName)
	}
	if stat.Size() > PortableMaxData {
		return item, errors.New("Datei zu groß")
	}
	w, err := zw.Create(archiveName)
	if err != nil {
		return item, err
	}
	h := sha256.New()
	n, err := io.Copy(w, io.TeeReader(file, h))
	if err != nil {
		return item, err
	}
	return PortableEntry{Path: archiveName, SHA256: hex.EncodeToString(h.Sum(nil)), Size: n}, nil
}

// CreatePortableZIP only runs under PortableDataLock.RLock to keep files and DB consistent.
func CreatePortableZIP() (string, PortableSummary, error) {
	var result PortableSummary
	config, assets, err := portablePaths()
	if err != nil {
		return "", result, err
	}
	tmpDir, err := os.MkdirTemp(config, ".hm-export-")
	if err != nil {
		return "", result, err
	}
	defer os.RemoveAll(tmpDir)
	snapshot := filepath.Join(tmpDir, "data.sqlite")
	if err = portableSnapshot(snapshot); err != nil {
		return "", result, err
	}
	schema, tables, err := portableSanitise(snapshot)
	if err != nil {
		return "", result, err
	}
	src, err := sql.Open("sqlite3", snapshot)
	if err != nil {
		return "", result, err
	}
	defer src.Close()
	counts, err := portableCount(src, tables)
	if err != nil {
		return "", result, err
	}
	names, err := portableAssets(assets)
	if err != nil {
		return "", result, err
	}
	available := map[string]bool{}
	for _, name := range names {
		available[name] = true
	}
	if err = portableCheckReferences(src, available); err != nil {
		return "", result, err
	}
	now := time.Now().UTC()
	m := PortableManifest{Format: portableFormat, FormatVersion: PortableBackupVersion,
		AppVersion: PortableAppVersion, SchemaVersion: PortableBackupVersion,
		SchemaSHA256: schema, CreatedAt: now, Counts: counts, Files: []PortableEntry{}}
	f, err := os.CreateTemp(config, ".hm-download-*.zip")
	if err != nil {
		return "", result, err
	}
	zipPath := f.Name()
	fail := func(e error) (string, PortableSummary, error) { f.Close(); os.Remove(zipPath); return "", result, e }
	if err = f.Chmod(0600); err != nil {
		return fail(err)
	}
	zw := zip.NewWriter(f)
	total := int64(0)
	add := func(name, path string) error {
		e, err := portableZipFile(zw, name, path)
		if err != nil {
			return err
		}
		total += e.Size
		if total > PortableMaxData {
			return errors.New("Backup überschreitet 1 GiB")
		}
		m.Files = append(m.Files, e)
		return nil
	}
	if err = add("data.sqlite", snapshot); err != nil {
		return fail(err)
	}
	for _, name := range names {
		if err = add("assets/"+name, filepath.Join(assets, name)); err != nil {
			return fail(err)
		}
	}
	manifestBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fail(err)
	}
	mw, err := zw.Create("manifest.json")
	if err != nil {
		return fail(err)
	}
	if _, err = mw.Write(manifestBytes); err != nil {
		return fail(err)
	}
	if err = zw.Close(); err != nil {
		return fail(err)
	}
	if err = f.Close(); err != nil {
		os.Remove(zipPath)
		return "", result, err
	}
	stat, err := os.Stat(zipPath)
	if err != nil {
		os.Remove(zipPath)
		return "", result, err
	}
	if stat.Size() > PortableMaxZIP {
		os.Remove(zipPath)
		return "", result, errors.New("ZIP größer als 512 MiB")
	}
	result = PortableSummary{Manifest: m, AssetCount: len(names),
		PasswordNotice: "Passwörter, Hashes, Tokens und .env sind nicht enthalten."}
	log.Printf("portable backup created: %d assets, %d database tables", len(names), len(tables))
	return zipPath, result, nil
}

func portableExtract(path, stage string) (PortableManifest, map[string]bool, error) {
	var m PortableManifest
	stat, err := os.Stat(path)
	if err != nil {
		return m, nil, err
	}
	if stat.Size() == 0 || stat.Size() > PortableMaxZIP {
		return m, nil, errors.New("ZIP ist leer oder größer als 512 MiB")
	}
	reader, err := zip.OpenReader(path)
	if err != nil {
		return m, nil, errors.New("Keine gültige ZIP-Datei")
	}
	defer reader.Close()
	if err = os.MkdirAll(filepath.Join(stage, "assets"), 0700); err != nil {
		return m, nil, err
	}
	if len(reader.File) < 2 || len(reader.File) > PortableMaxFiles+2 {
		return m, nil, errors.New("Ungültige Anzahl Dateien im ZIP")
	}
	seen := map[string]PortableEntry{}
	total := int64(0)
	var manifestBytes []byte
	for _, entry := range reader.File {
		name := entry.Name
		if !entry.Mode().IsRegular() && entry.Mode()&os.ModeType != 0 {
			return m, nil, errors.New("Nur reguläre ZIP-Dateien erlaubt")
		}
		if name != "manifest.json" && name != "data.sqlite" &&
			!(strings.HasPrefix(name, "assets/") && portableAssetName(strings.TrimPrefix(name, "assets/"))) {
			return m, nil, errors.New("Unerwarteter oder unsicherer ZIP-Pfad")
		}
		if _, exists := seen[name]; exists {
			return m, nil, errors.New("Doppelter ZIP-Eintrag")
		}
		if entry.UncompressedSize64 > uint64(PortableMaxData) || total+int64(entry.UncompressedSize64) > PortableMaxData {
			return m, nil, errors.New("ZIP enthält zu viele oder zu große Daten")
		}
		total += int64(entry.UncompressedSize64)
		input, err := entry.Open()
		if err != nil {
			return m, nil, err
		}
		var out *os.File
		if name == "manifest.json" {
			if entry.UncompressedSize64 > 1<<20 {
				input.Close()
				return m, nil, errors.New("Manifest zu groß")
			}
			data, readErr := io.ReadAll(io.LimitReader(input, 1<<20+1))
			input.Close()
			if readErr != nil {
				return m, nil, readErr
			}
			if len(data) > 1<<20 {
				return m, nil, errors.New("Manifest zu groß")
			}
			manifestBytes = data
			seen[name] = PortableEntry{Path: name}
			continue
		}
		dest := filepath.Join(stage, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			input.Close()
			return m, nil, err
		}
		out, err = os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			input.Close()
			return m, nil, err
		}
		h := sha256.New()
		n, readErr := io.Copy(out, io.TeeReader(io.LimitReader(input, PortableMaxData+1), h))
		closeErr := out.Close()
		input.Close()
		if readErr != nil {
			return m, nil, readErr
		}
		if closeErr != nil {
			return m, nil, closeErr
		}
		if n != int64(entry.UncompressedSize64) || n > PortableMaxData {
			return m, nil, errors.New("Ungültige ZIP-Dateigröße")
		}
		seen[name] = PortableEntry{Path: name, SHA256: hex.EncodeToString(h.Sum(nil)), Size: n}
	}
	if err = json.Unmarshal(manifestBytes, &m); err != nil {
		return m, nil, errors.New("Ungültiges Backup-Manifest")
	}
	if m.Format != portableFormat || m.FormatVersion != PortableBackupVersion ||
		m.SchemaVersion != PortableBackupVersion || m.AppVersion == "" ||
		m.CreatedAt.IsZero() || len(m.Files) != len(seen)-1 {
		return m, nil, errors.New("Nicht unterstütztes Backup-Format oder Schema")
	}
	expected := map[string]bool{}
	for _, e := range m.Files {
		if expected[e.Path] || e.Path == "manifest.json" {
			return m, nil, errors.New("Doppelter Manifest-Eintrag")
		}
		actual, ok := seen[e.Path]
		if !ok || actual.SHA256 != e.SHA256 || actual.Size != e.Size {
			return m, nil, errors.New("Backup beschädigt: Prüfsumme oder Dateigröße stimmt nicht")
		}
		expected[e.Path] = true
	}
	if !expected["data.sqlite"] {
		return m, nil, errors.New("SQLite-Datenbank fehlt im Backup")
	}
	assets := map[string]bool{}
	for name := range expected {
		if strings.HasPrefix(name, "assets/") {
			assets[strings.TrimPrefix(name, "assets/")] = true
		}
	}
	return m, assets, nil
}

func portableCheckRestore(stage string, m PortableManifest, files map[string]bool, admin db.User) (PortableSummary, error) {
	result := PortableSummary{Manifest: m, AssetCount: len(files), PasswordNotice: "Das Passwort des Ziel-Administrators bleibt erhalten; weitere Benutzer müssen neue Passwörter erhalten."}
	incoming, err := sql.Open("sqlite3", "file:"+filepath.Join(stage, "data.sqlite")+"?mode=ro")
	if err != nil {
		return result, err
	}
	defer incoming.Close()
	if err = portableCheckSQLite(incoming); err != nil {
		return result, err
	}
	hash, tables, err := portableSchema(incoming)
	if err != nil {
		return result, err
	}
	if hash != m.SchemaSHA256 {
		return result, errors.New("Backup-Schema stimmt nicht mit dem Manifest überein")
	}
	sqlDB, err := db.DB.DB()
	if err != nil {
		return result, err
	}
	liveHash, liveTables, err := portableSchema(sqlDB)
	if err != nil {
		return result, err
	}
	if len(m.Counts) != len(tables) {
		return result, errors.New("Datensatztabellen im Manifest stimmen nicht")
	}
	for _, table := range tables {
		if _, exists := m.Counts[table]; !exists { return result, errors.New("Datensatztabelle im Manifest fehlt") }
	}
	if liveHash != hash || len(liveTables) != len(tables) {
		// v1.1.6 lacks tires and mounting states; v1.1.7 lacks states.
		// Every other table/column must match exactly.
		omit:=map[string]bool{}
		for _,table:=range liveTables{
			if _,present:=m.Counts[table];!present{
				if table!="tires"&&table!="tire_mileage_states"&&table!="maintenance_rules"&&table!="recurring_costs"&&table!="financing_terms"&&table!="financing_calculations"&&table!="home_assistant_keys"{return result,errors.New("Backup und Zielinstallation haben unterschiedliche Datenbankschemas")}
				omit[table]=true
			}
		}
		_,hasTires:=m.Counts["tires"];_,hasStates:=m.Counts["tire_mileage_states"]
		legacyHash,legacyTables,legacyErr:=portableSchemaOmitting(sqlDB,omit)
		if (!hasTires&&hasStates)||len(liveTables)!=len(tables)+len(omit)||legacyErr!=nil||legacyHash!=hash||len(legacyTables)!=len(tables){
			return result,errors.New("Backup und Zielinstallation haben unterschiedliche Datenbankschemas. Bitte dieselbe App-Version einsetzen.")
		}
	}
	for _, table := range tables {
		var value int64
		if err = incoming.QueryRow(`SELECT COUNT(*) FROM "` + table + `"`).Scan(&value); err != nil {
			return result, err
		}
		if value != m.Counts[table] {
			return result, errors.New("Datensatzanzahl im Backup stimmt nicht")
		}
	}
	if _,present:=m.Counts["tire_mileage_states"];present{
		var invalid int
		if err=incoming.QueryRow("SELECT COUNT(*) FROM tire_mileage_states s LEFT JOIN tires t ON t.id=s.tire_id WHERE t.id IS NULL OR t.status <> 'mounted' OR t.position = 'spare' OR s.mount_odo IS NULL OR s.mount_odo < 0 OR s.mount_odo > 10000000 OR s.mounted_on IS NULL OR date(s.mounted_on) IS NULL OR s.mounted_on <> date(s.mounted_on)").Scan(&invalid);err!=nil{return result,err}
		if invalid!=0{return result,errors.New("Ungültige Reifen-Montagedaten im Backup")}
	}
	if _,present:=m.Counts["maintenance_rules"];present{
		var invalid int
		if err=incoming.QueryRow("SELECT COUNT(*) FROM maintenance_rules r LEFT JOIN maintenance_tasks t ON t.id=r.task_id WHERE t.id IS NULL OR r.repeat_months IS NULL OR r.repeat_months NOT BETWEEN 0 AND 120 OR r.repeat_km IS NULL OR r.repeat_km NOT BETWEEN 0 AND 1000000 OR r.reminder_days IS NULL OR r.reminder_days NOT BETWEEN 0 AND 365 OR r.reminder_km IS NULL OR r.reminder_km NOT BETWEEN 0 AND 1000000 OR r.notification_stage IS NULL OR r.notification_stage NOT BETWEEN 0 AND 2 OR (r.completed_odo IS NOT NULL AND r.completed_odo NOT BETWEEN 0 AND 10000000)").Scan(&invalid);err!=nil{return result,err}
		if invalid!=0{return result,errors.New("Ungültige Wartungsregeln im Backup")}
	}
	if _,present:=m.Counts["recurring_costs"];present{
		var invalid int
		if err=incoming.QueryRow("SELECT COUNT(*) FROM recurring_costs WHERE amount_cents IS NULL OR amount_cents NOT BETWEEN 1 AND 100000000 OR interval_months IS NULL OR interval_months NOT IN (1,3,6,12) OR installments IS NULL OR installments NOT BETWEEN 0 AND 600 OR generated IS NULL OR generated NOT BETWEEN 0 AND 12000 OR (installments > 0 AND generated > installments) OR booked_cents IS NULL OR booked_cents < 0 OR start_date IS NULL OR date(start_date) IS NULL OR start_date <> date(start_date) OR (final_cents IS NOT NULL AND (installments = 0 OR final_cents NOT BETWEEN 1 AND 100000000))").Scan(&invalid);err!=nil{return result,err}
		if invalid!=0{return result,errors.New("Ungültige Kostenpläne im Backup")}
	}
	if _,present:=m.Counts["financing_terms"];present{
		var invalid int
		if err=incoming.QueryRow("SELECT COUNT(*) FROM financing_terms f LEFT JOIN recurring_costs p ON p.id=f.plan_id WHERE p.id IS NULL OR p.category <> 'Finanzierung' OR f.down_payment_cents IS NULL OR f.down_payment_cents NOT BETWEEN 0 AND 100000000 OR f.already_recorded IS NULL OR f.already_recorded NOT IN (0,1) OR (f.down_payment_cents > 0 AND (f.down_payment_date IS NULL OR date(f.down_payment_date) IS NULL OR f.down_payment_date <> date(f.down_payment_date) OR f.down_payment_date < '1900-01-01')) OR (f.down_payment_cents = 0 AND (COALESCE(f.down_payment_date,'') <> '' OR f.already_recorded <> 0)) OR (f.principal_cents IS NOT NULL AND f.principal_cents NOT BETWEEN 1 AND 100000000) OR (f.nominal_bps IS NOT NULL AND f.nominal_bps NOT BETWEEN 0 AND 10000) OR (f.effective_bps IS NOT NULL AND f.effective_bps NOT BETWEEN 0 AND 10000)").Scan(&invalid);err!=nil{return result,err}
		if invalid!=0{return result,errors.New("Ungültige Finanzierungsdetails im Backup")}
	}
	if _,present:=m.Counts["financing_calculations"];present{
		var invalid int
		if err=incoming.QueryRow("SELECT COUNT(*) FROM financing_calculations c LEFT JOIN financing_terms t ON t.plan_id=c.plan_id WHERE t.plan_id IS NULL OR (c.monthly_rate_ppm IS NULL AND c.first_interest_cents IS NULL) OR (c.monthly_rate_ppm IS NOT NULL AND c.monthly_rate_ppm NOT BETWEEN 0 AND 100000) OR (c.first_interest_cents IS NOT NULL AND c.first_interest_cents NOT BETWEEN 0 AND 100000000)").Scan(&invalid);err!=nil{return result,err}
		if invalid!=0{return result,errors.New("Ungültige Zinsberechnungsdaten im Backup")}
	}
	if _,present:=m.Counts["home_assistant_keys"];present{
		var count int
		if err=incoming.QueryRow("SELECT COUNT(*) FROM home_assistant_keys").Scan(&count);err!=nil{return result,err}
		if count!=0{return result,errors.New("Ein Backup darf keine Home-Assistant-Zugangsschlüssel enthalten")}
	}
	if err = portableCheckReferences(incoming, files); err != nil {
		return result, err
	}
	var count int
	err = incoming.QueryRow("SELECT COUNT(*) FROM users WHERE lower(email)=lower(?) AND role=?",
		admin.Email, int(db.ADMIN)).Scan(&count)
	if err != nil {
		return result, err
	}
	if count != 1 {
		return result, errors.New("Im Backup fehlt ein Admin mit der E-Mail des aktuellen Ziel-Admins. Bitte auf dem Zielserver zuerst einen Administrator mit derselben E-Mail wie im Backup anlegen.")
	}
	result.AdminEmailMatched = true
	return result, nil
}

// InspectPortableZIP fully extracts, hashes, validates and checks schema; never changes live data.
func InspectPortableZIP(path string, admin db.User) (PortableSummary, error) {
	var empty PortableSummary
	dir, err := os.MkdirTemp("", "hm-inspect-")
	if err != nil {
		return empty, err
	}
	defer os.RemoveAll(dir)
	m, files, err := portableExtract(path, dir)
	if err != nil {
		return empty, err
	}
	return portableCheckRestore(dir, m, files, admin)
}

func portableCopy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	stat, err := in.Stat()
	if err != nil {
		return err
	}
	if !stat.Mode().IsRegular() {
		return errors.New("Ungültiger Dateityp im Backup")
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	return err
}

func portableRescue(config, assets string) (string, error) {
	folder := filepath.Join(config, "restore-safety")
	if err := os.MkdirAll(folder, 0700); err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp(folder, "pre-restore-")
	if err != nil {
		return "", err
	}
	// A local rescue is deliberately UNSANITISED: it preserves existing logins.
	// It is never exposed by the download API.
	if err = portableSnapshot(filepath.Join(dir, "hammond.db")); err != nil {
		return "", err
	}
	if err = os.Mkdir(filepath.Join(dir, "assets"), 0700); err != nil {
		return "", err
	}
	names, err := portableAssets(assets)
	if err != nil {
		return "", err
	}
	for _, name := range names {
		if err = portableCopy(filepath.Join(assets, name), filepath.Join(dir, "assets", name)); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// Verify live data before the recovery journal is removed. A failed check
// leaves the journal available so startup can restore the previous state.
func portableVerifyRestored(conn *sql.DB, assets string, manifest PortableManifest) error {
	if err := portableCheckSQLite(conn); err != nil {
		return err
	}
	for table, expected := range manifest.Counts {
		if !allowedTables[table] || !safeTable.MatchString(table) {
			return errors.New("Unerwartete Tabelle bei Abschlussprüfung")
		}
		var count int64
		if err := conn.QueryRow(`SELECT COUNT(*) FROM "` + table + `"`).Scan(&count); err != nil {
			return err
		}
		if count != expected {
			return fmt.Errorf("Datensätze fehlen nach Wiederherstellung: %s", table)
		}
	}
	for _, item := range manifest.Files {
		if !strings.HasPrefix(item.Path, "assets/") {
			continue
		}
		name := strings.TrimPrefix(item.Path, "assets/")
		if !portableAssetName(name) {
			return errors.New("Ungültiger Anhang nach Wiederherstellung")
		}
		stat, err := os.Lstat(filepath.Join(assets, name))
		if err != nil {
			return err
		}
		if !stat.Mode().IsRegular() || stat.Size() != item.Size {
			return errors.New("Dateien nach Wiederherstellung unvollständig")
		}
	}
	return nil
}

func portableSwapAssets(assets, stage string) (func() error, func() error, error) {
	old, err := os.MkdirTemp(assets, ".hm-rollback-")
	if err != nil {
		return nil, nil, err
	}
	movedOld := []string{}
	movedNew := []string{}
	rollback := func() error {
		var failures []string
		for _, name := range movedNew {
			if err := os.Remove(filepath.Join(assets, name)); err != nil {
				failures = append(failures, err.Error())
			}
		}
		for _, name := range movedOld {
			if err := os.Rename(filepath.Join(old, name), filepath.Join(assets, name)); err != nil {
				failures = append(failures, err.Error())
			}
		}
		if len(failures) == 0 {
			os.RemoveAll(old)
		}
		if len(failures) > 0 {
			return errors.New(strings.Join(failures, "; "))
		}
		return nil
	}
	names, err := portableAssets(assets)
	if err != nil {
		return nil, nil, err
	}
	for _, name := range names {
		if err = os.Rename(filepath.Join(assets, name), filepath.Join(old, name)); err != nil {
			_ = rollback()
			return nil, nil, err
		}
		movedOld = append(movedOld, name)
	}
	replacement, err := portableAssets(stage)
	if err != nil {
		_ = rollback()
		return nil, nil, err
	}
	for _, name := range replacement {
		if err = os.Rename(filepath.Join(stage, name), filepath.Join(assets, name)); err != nil {
			_ = rollback()
			return nil, nil, err
		}
		movedNew = append(movedNew, name)
	}
	cleanup := func() error { return os.RemoveAll(old) }
	return rollback, cleanup, nil
}

// RestorePortableZIP must be called while holding PortableDataLock.Lock.
// The live SQLite replacement is one transaction. File changes can be rolled
// back on errors; a server-side rescue snapshot is kept for disaster recovery.
func RestorePortableZIP(path string, admin db.User) (PortableSummary, error) {
	var empty PortableSummary
	config, assets, err := portablePaths()
	if err != nil {
		return empty, err
	}
	stage, err := os.MkdirTemp(assets, ".hm-stage-")
	if err != nil {
		return empty, err
	}
	defer os.RemoveAll(stage)
	m, files, err := portableExtract(path, stage)
	if err != nil {
		return empty, err
	}
	result, err := portableCheckRestore(stage, m, files, admin)
	if err != nil {
		return result, err
	}
	rescue, err := portableRescue(config, assets)
	if err != nil {
		return result, fmt.Errorf("Sicherheits-Backup konnte nicht erstellt werden: %w", err)
	}
	result.SafetyBackup = rescue
	log.Printf("portable restore: safety backup saved in %s", rescue)
	sqlDB, err := db.DB.DB()
	if err != nil {
		return result, err
	}
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		return result, err
	}
	defer conn.Close()
	if _, err = conn.ExecContext(context.Background(), "PRAGMA foreign_keys=OFF"); err != nil {
		return result, err
	}
	defer conn.ExecContext(context.Background(), "PRAGMA foreign_keys=ON")
	if _, err = conn.ExecContext(context.Background(), "ATTACH DATABASE ? AS incoming", filepath.Join(stage, "data.sqlite")); err != nil {
		return result, err
	}
	defer conn.ExecContext(context.Background(), "DETACH DATABASE incoming")
	_, tables, err := portableSchema(sqlDB)
	if err != nil {
		return result, err
	}
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	for i := len(tables) - 1; i >= 0; i-- {
		if _, err = tx.Exec(`DELETE FROM main."` + tables[i] + `"`); err != nil {
			return result, err
		}
	}
	for _, table := range tables {
		if table == "tires" || table == "tire_mileage_states" || table == "maintenance_rules" || table == "recurring_costs" || table == "financing_terms" || table == "financing_calculations" || table == "home_assistant_keys" { if _, present := m.Counts[table]; !present { continue } }
		if _, err = tx.Exec(`INSERT INTO main."` + table + `" SELECT * FROM incoming."` + table + `"`); err != nil {
			return result, err
		}
	}
	// Imported account hashes are blank; preserve the currently authenticated
	// destination admin's credential but keep the original user's stable ID.
	if _, err = tx.Exec("UPDATE main.users SET password=?, is_disabled=0 WHERE lower(email)=lower(?) AND role=?",
		admin.Password, admin.Email, int(db.ADMIN)); err != nil {
		return result, err
	}
	// Attachment paths use /assets inside the container. Never trust paths
	// from the imported SQLite database.
	rows, err := tx.Query("SELECT id,path FROM main.attachments WHERE deleted_at IS NULL")
	if err != nil {
		return result, err
	}
	refs := map[string]string{}
	for rows.Next() {
		var id, p string
		if err = rows.Scan(&id, &p); err != nil {
			break
		}
		refs[id] = filepath.Base(p)
	}
	if e := rows.Close(); err == nil {
		err = e
	}
	if err != nil {
		return result, err
	}
	for id, name := range refs {
		if !portableAssetName(name) || !files[name] {
			return result, errors.New("Fehlender Anhang in der Wiederherstellung")
		}
		if _, err = tx.Exec("UPDATE main.attachments SET path=? WHERE id=?", filepath.Join(assets, name), id); err != nil {
			return result, err
		}
	}
	// The journal is durable BEFORE the first file move; after an unplanned
	// container restart, RecoverPortableRestore reverts to the rescue snapshot.
	if err = portableStartJournal(config, rescue); err != nil {
		return result, err
	}
	rollbackFiles, cleanupFiles, err := portableSwapAssets(assets, filepath.Join(stage, "assets"))
	if err != nil {
		return result, fmt.Errorf("Dateitausch fehlgeschlagen; vor neuer Bearbeitung Container für automatische Recovery neu starten: %w", err)
	}
	if err = tx.Commit(); err != nil {
		if rollbackErr := rollbackFiles(); rollbackErr != nil {
			log.Printf("CRITICAL restore file rollback failed: %v; safety snapshot: %s", rollbackErr, rescue)
			return result, errors.New("Datenbankfehler und Rollback-Fehler: Container zur automatischen Recovery neu starten")
		}
		if journalErr := portableFinishJournal(config); journalErr != nil {
			log.Printf("restore journal cleanup failed after rollback: %v", journalErr)
		}
		return result, err
	}
	if err = cleanupFiles(); err != nil {
		log.Printf("portable restore: old assets cleanup: %v", err)
	}
	if err = portableVerifyRestored(sqlDB, assets, m); err != nil {
		log.Printf("CRITICAL restore verification failed: %v; recovery journal retained", err)
		return result, errors.New("Wiederherstellung konnte nicht bestätigt werden; Container für automatische Recovery neu starten")
	}
	if err = portableFinishJournal(config); err != nil {
		log.Printf("CRITICAL restore journal removal failed: %v", err)
		return result, errors.New("Wiederherstellung durchgeführt, aber Recovery-Journal konnte nicht entfernt werden; bitte Serverzugriff prüfen")
	}
	log.Printf("portable restore completed: %d assets, %d tables", len(files), len(tables))
	return result, nil
}
