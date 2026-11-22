# Backup, Wiederherstellung und Serverumzug – Schritt für Schritt

> **Stable v1.1.8:** ZIP-Export und Import wurden praktisch getestet. Vor einem produktiven Umzug externe Sicherung anlegen. Ältere Versionen zuerst unter Beibehaltung der bisherigen Datenpfade aktualisieren.

**Ab v1.1.7:** Reifen einschließlich Fahrzeugzuordnungen werden zusätzlich gesichert. Ältere ZIP-Backups ohne Reifentabelle werden akzeptiert, wenn alle übrigen Tabellen und Spalten exakt passen. Ein Restore ersetzt auch bestehende Reifen – ein altes Backup ohne Reifentabelle ergibt eine leere Reifenliste. Ein neuer Reifen-ZIP kann nicht in eine unveränderte 1.1.6-Instanz importiert werden. Vor Umstieg und Rückkehr immer externe Sicherung.

**Ab v1.1.8:** Aktive Montageintervalle werden ebenfalls gesichert. Kompatible ZIPs aus v1.1.7 behalten die gespeicherte bisherige Reifenlaufleistung, enthalten jedoch keine aktiven Montage-Kilometerstände; diese müssen bei vorhandenen montierten Reifen neu hinterlegt werden. ZIPs aus v1.1.8 mit der neuen Tabelle können nicht in unveränderte ältere Versionen importiert werden.

## Was wird gesichert?

| Bestandteil | Im portablen ZIP? |
| --- | --- |
| Fahrzeuge und Stammdaten | Ja |
| Tankungen und Kilometerstände | Ja |
| Ausgaben und Werkstattkosten | Ja |
| Wartungen, Fahrer, Zuordnungen und Freigaben | Ja |
| Reifen und Fahrzeugzuordnungen | Ja (ab v1.1.7) |
| Wartungsregeln und wiederkehrende Kosten inkl. Buchungsfortschritt | Ja (ab 1.1.9-dev); ältere ZIPs enthalten diese neuen Tabellen nicht |
| Aktive Montageintervalle und bisherige Reifenlaufleistung | Ja (ab v1.1.8); ältere ZIPs ohne Intervalle benötigen erneuten Startstand |
| Rechnungen und zulässige Uploads aus `/assets` | Ja |
| Benutzerprofile und Rollen | Ja |
| Währung, Einheit, Datumsformat und Einstellungen aus SQLite | Ja |
| SQLite-Schema und gespeicherte Migrationsstände | Ja |
| Exportzeit, App-Version, Schema-Fingerabdruck, SHA-256-Prüfsummen, Datensatzanzahl | Ja |
| Passwörter, **auch Passwort-Hashes** | **Nein** |
| `.env`, `JWT_SECRET`, API-Keys, Tokens | **Nein** |
| Docker Compose, Docker-Images, Netzwerke, Serverkonfiguration | **Nein** |
| Dateien außerhalb von `/assets` und `/config/hammond.db` | **Nein** |

Das ZIP enthält die **bereinigte SQLite-Datenbank** und reguläre Dateien aus dem
`/assets`-Verzeichnis, mit Manifest. Es ist **nicht verschlüsselt**:
Fahrzeughistorien, E-Mail-Adressen und Rechnungen können persönliche Daten enthalten.
ZIP-Dateien sicher verwahren und nicht öffentlich hochladen.

Unterstützte Dateitypen für Anhänge:
PDF, JPG/JPEG, PNG, WEBP, GIF, BMP, TIFF, HEIC, TXT, CSV, DOC/DOCX, XLS/XLSX.
Andere Dateitypen oder Verzeichnisse unter `/assets` führen aus Sicherheitsgründen
zu einer verständlichen Fehlermeldung, statt unbemerkt ignoriert zu werden.
Beschädigte oder fehlende Anhänge werden ebenfalls gemeldet.

## Wichtiger Unterschied: Backup der App und Docker-Konfiguration

- **Anwendung:** Quellcode oder Docker-Image; auf dem neuen Host neu beziehen bzw. bauen.
- **Konfiguration:** `.env`, eigene Secrets, Host-Port, Docker-Compose-Datei und
  externe Netzwerke – getrennt verwalten. Sie werden **nicht** aus dem ZIP importiert.
- **Persistente Daten:** SQLite im Container unter `/config/hammond.db` und
  hochgeladene Dateien unter `/assets`. Diese Daten werden vom ZIP gesichert.

`docker export` und `docker save` sichern **keine Volume-Daten**.
Die portable App-Sicherung ersetzt keine separate sicher gespeicherte Kopie der
ursprünglichen Datenverzeichnisse.

## A. Frische Installation mit benannten Docker-Volumes

Für eine neue Installation des festen Releases:

```bash
git clone --branch v1.1.8 https://github.com/Skoun181/hammond-modern.git
cd hammond-modern
mkdir -p data
(umask 077; printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)" > .env)
docker compose up -d --build
docker compose ps
```

Diese **Compose-Datei** benutzt zwei benannte Docker-Volumes:
`hammond_config` für `/config` und `hammond_assets` für `/assets`.
Docker vergibt gewöhnlich Namen mit dem Compose-Projektpräfix, z. B.
`hammond-modern_hammond_config`. Den tatsächlichen Namen bei Bedarf
mit `docker volume ls` prüfen.

Die App ist standardmäßig erreichbar unter `http://SERVER-IP:3006`.
Bei einem anderen freien externen Port in `.env` eintragen:

```dotenv
HAMMOND_PORT=8085
```

Dann `docker compose up -d` ausführen. Zugriff unter
`http://SERVER-IP:8085`. **Der interne Port bleibt 3000.**

**Achtung für bestehende Synology-Installationen:** Wenn deine Daten bisher
über die Bind Mounts `/volume2/docker/Hammond/Hammond-config` und
`/volume2/docker/Hammond/Hammond-assets` bereitgestellt sind, liegen sie
**nicht automatisch in den neuen benannten Volumes**. Die alten Mounts bitte
nicht entfernen, bevor ein getestetes externes Backup existiert.

## B. Export auf dem alten Rechner

1. Bei Hammond Modern **als Administrator** anmelden.
2. Links im Menü **Backup & Umzug** auswählen.
3. Auf **Backup herunterladen (.zip)** klicken.
4. Der Browser lädt etwa `hammond-modern-backup-2026-11-20-183000.zip` herunter.
5. Die ZIP-Datei sicher auf dem eigenen Rechner oder einem separaten Backup-Medium ablegen.
6. Für zusätzliche Sicherheit den alten Container stoppen und die alten
   `/config`- und `/assets`-Verzeichnisse separat sichern.

Während der Erstellung wird die Anwendung gegen konkurrierende HTTP-Schreibzugriffe
gesperrt; SQLite wird mittels `VACUUM INTO` konsistent kopiert.

**Nicht vergessen:** `.env`, Docker-Compose-Konfiguration, Reverse-Proxy-Einstellungen,
externe Netzwerke und eigenen Tankerkönig-API-Key separat sichern bzw. neu einrichten.
Niemals Secrets öffentlich teilen.

## C. Neue Docker-Installation vorbereiten

1. Docker und Docker Compose auf dem Zielserver installieren.
2. Hammond Modern **v1.1.8** bzw. dieselbe kompatible Version bereitstellen (Abschnitt A oder Docker-Hub-Installation in der README). Ältere Container ohne Backup-Funktion können das ZIP nicht importieren.
3. Eigene `.env` mit **neuem** zufälligem `JWT_SECRET` erstellen. Andere API-Keys
   nur bei Bedarf erneut eintragen.
4. Lokaler Build: `docker compose up -d --build`. Docker Hub: `docker compose pull && docker compose up -d`.
5. Hammond Modern im Browser öffnen.
6. **Wichtig:** Den ersten Administrator auf dem neuen Server **mit genau derselben
   E-Mail-Adresse wie einen Administrator der alten Installation** anlegen.
   Ein **neues** Passwort ist möglich – dieses bleibt beim Restore erhalten.
7. Im Anschluss als dieser neue Administrator anmelden.

Es müssen keine Fahrzeuge oder Tankungen manuell angelegt werden.

## D. ZIP auf dem neuen Server importieren

1. Als Ziel-Administrator **Backup & Umzug** öffnen.
2. ZIP-Datei von der alten Installation auswählen.
3. **Backup prüfen** anklicken. Der Server prüft:
   - ZIP-Dateigröße (maximal 512 MiB) und entpackte Größe (maximal 1 GiB)
   - erlaubte Dateien und Pfade, keine ZIP-Symlinks oder doppelten Namen
   - CRC, SHA-256, SQLite-Integrität und Datensatzanzahlen
   - Backup-Format, Schema-Fingerabdruck und kompatible Datenbanktabellen
   - passende Admin-E-Mail-Adresse auf der Zielinstallation
4. Die angezeigte Vorschau (Fahrzeuge, Tankungen, Ausgaben, Wartungen und Dateien)
   kontrollieren.
5. **Achtung: Der Import ersetzt vorhandene Daten!** Warnung und Bestätigung lesen,
   Checkbox aktivieren und `ERSETZEN` eintippen.
6. **Daten endgültig wiederherstellen** klicken.
7. Nach Erfolg wird die Anzahl der übernommenen Datensätze und Dateien angezeigt.
8. Abmelden und mit der E-Mail-Adresse des Ziel-Administrators und dessen
   **bisherigem Ziel-Passwort** neu anmelden.
9. Prüfen: Fahrzeugliste, Tankungen, Kosten, Wartungstermine und Rechnungsdownloads.

**Benutzer außer dem angemeldeten Administrator** haben im neuen System
zunächst keine gültigen Passwörter. Unter **Benutzer** kann der Administrator
für jedes dieser Konten ein neues Passwort (mindestens 12 Zeichen) festlegen.
Deaktivierte Konten müssen bei Bedarf danach wieder aktiviert werden.

## E. Was passiert beim Wiederherstellen?

Die Anwendung kopiert zuerst die aktuell vorhandene SQLite-Datenbank
**unbereinigt** und alle bisherigen Anhänge in den geschützten Ordner
`/config/restore-safety/pre-restore-*/`. Darin können Passwort-Hashes
der bisherigen Installation enthalten sein! Dies ist eine **lokale**
Rettungskopie, die nicht über die Export-API ausgeliefert wird.
Zugriff auf das Volume entsprechend einschränken.

Die eingelesene Datenbank wird innerhalb einer SQLite-Transaktion ersetzt.
Anhänge werden in einen Staging-Bereich entpackt und vor dem Commit
innerhalb des Daten-Volumes umgeschaltet. Bei normalen Fehlern versucht die
Anwendung, die bisherigen Anhänge zurückzurollen. Das Sicherheitsbackup
bleibt unabhängig davon zur manuellen Wiederherstellung erhalten.

**Neustart nach Unterbrechung:** Vor dem ersten Öffnen von SQLite prüft
Hammond Modern ein dauerhaftes Recovery-Journal. Wird darin ein
unterbrochener Import gefunden, stellt die Anwendung die **alte**
Datenbank und Uploads aus der Rettungskopie wieder her. Kann die
Wiederherstellung nicht durchgeführt werden (z. B. beschädigtes Volume),
**startet der Server absichtlich nicht**, statt möglicherweise halbfertige
Daten freizugeben. Das Journal muss nach einer erfolgreichen Recovery
verschwunden sein.

Ein Stromausfall oder Plattenfehler kann trotzdem sowohl das Daten- als
auch das Rettungsvolume beschädigen. Deshalb **immer auch eine externe
Offline-Sicherung** anlegen und Restore zunächst in einem separaten
Testcontainer üben.

## F. Separates Backup benannter Volumes (Notfallweg)

Falls die Weboberfläche nicht mehr funktioniert, lassen sich benannte Volumes
zusätzlich offline als TAR-Datei sichern. Den tatsächlichen Volume-Namen
über `docker volume ls` ermitteln. Beispiel:

```bash
docker compose stop

docker run --rm \
  -v hammond-modern_hammond_config:/data:ro \
  -v "$PWD":/backup \
  alpine sh -c 'tar czf /backup/config.tar.gz -C /data .'

docker run --rm \
  -v hammond-modern_hammond_assets:/data:ro \
  -v "$PWD":/backup \
  alpine sh -c 'tar czf /backup/assets.tar.gz -C /data .'

docker compose start
```

Diese **rohen Volumesicherungen** enthalten alle Daten, einschließlich
Passwort-Hashes aus SQLite! Sie sind nicht dasselbe wie das bereinigte,
portable ZIP. Sicher und verschlüsselt speichern.

Für eine manuelle Rücksicherung den Container stoppen, die
TAR-Dateien **nur aus einer eigenen vertrauenswürdigen Sicherung**
in die entsprechenden Volumes zurückspielen und den Container starten.
Archive fremder Herkunft nicht ungeprüft als Root entpacken.

## G. Weitere Hinweise

- Vorher vorhandene Daten in der Zielinstallation werden ersetzt, nicht zusammengeführt.
- Export und Restore sind **nur für Administratoren** freigeschaltet.
- Weitere SQLite-Tabellen oder abweichende Schemas werden vorerst aus
  Sicherheitsgründen abgewiesen, statt unkontrolliert übernommen.
- ZIP-Backups sind unverschlüsselt und sollten nicht per ungesichertem HTTP
  über das Internet übertragen werden. Reverse Proxy mit HTTPS oder VPN verwenden.
- Regelmäßig Backup **und Wiederherstellung** testen.
- Bestehende Release-Tags bleiben unverändert. Für die Backup-Funktion v1.1.6 oder eine ausdrücklich kompatible neuere Version verwenden.
