# Hammond Modern 🚗

**Self-Hosted-Fahrzeugverwaltung mit modernisierter Oberfläche, Wartungsplanung und einfacher Datenübernahme aus dem ursprünglichen Hammond.**

Hammond Modern ist ein unabhängiger Community-Fork von [akhilrex/hammond](https://github.com/akhilrex/hammond). Das Projekt ist **nicht offiziell mit dem Originalautor verbunden**.

**Version: v1.1.6** · [Releases](https://github.com/Skoun181/hammond-modern/releases) · [Änderungsverlauf](CHANGELOG-MODERN.md)

> **Teststand:** Die alltäglichen Funktionen und die Migrationsvorschau wurden auf einer privaten Debian-13-LXC-Installation getestet. Das Docker-Hub-Image wurde außerdem erfolgreich auf einer Synology über Dockhand und macvlan gestartet (Anwenderbestätigung). Der portable ZIP-Export und -Import wurden praktisch getestet (Anwenderbestätigung). Ein vollständiger Original-Hammond-Import und umfassender Sicherheits- und Produktionstest sind weiterhin nicht bestätigt.

## Was kann Hammond Modern?

| Bereich | Funktionen |
| --- | --- |
| **Fahrzeuge** | Mehrere Fahrzeuge, Hersteller- und Modellsuche, Kennzeichen, Fahrzeugfreigaben |
| **Tankungen** | Menge, Kraftstoff, Kilometerstand, Preis, Gesamtbetrag, Tankstelle und Fahrer |
| **Tankstellensuche** | Suche per deutscher PLZ mit eigenem Tankerkönig-API-Key |
| **Kosten** | Tankkosten, Werkstattkosten, sonstige Ausgaben, Rechnungen und historische Statistiken |
| **Wartung** | TÜV/HU, Inspektion und Service nach Datum und/oder Kilometerstand planen |
| **Dashboard** | Fahrzeugübersicht, anstehende Wartungen, Kostenübersicht und Verbrauch |
| **Backup & Umzug** | Portable ZIP-Sicherung mit Datenbank, Rechnungen, Uploads und geführter Wiederherstellung |
| **Datenmigration** | Original-Hammond-SQLite-Datei prüfen; Fahrzeuge gezielt auswählen und importieren |
| **Oberfläche** | Responsives Layout für PC und Smartphone, Dark-/Light-Mode, deutsche Beschriftungen |

Die Verbrauchsberechnung benötigt mindestens zwei geeignete Volltankungen mit aufsteigenden Kilometerständen. NHTSA-Modellvorschläge decken nicht alle europäischen Fahrzeuge ab; Modelle lassen sich manuell eingeben.

## Was ist neu in v1.1.6?

- Portables ZIP-Backup und Wiederherstellung für Administratoren, inklusive SQLite-Datenbank, Rechnungen und Uploads.
- Integritäts- und Schema-Prüfung, Vorschau, Sicherheitsbackup und Recovery-Journal vor Datenersetzung.
- Keine Passwort-Hashes oder Secrets im portablen ZIP. Ziel-Admin behält sein Passwort, weitere Benutzer benötigen neue Passwörter.
- Export und Import wurden praktisch getestet.

## Änderungen aus v1.1.5

### 1. Import von Original-Hammond
- Neuer Assistent unter **Datenimport → Hammond-Migration starten**.
- Upload einer konsistenten **`hammond.db`-Sicherung** aus der ursprünglichen Hammond-Installation (max. 50 MB).
- **Vorschau** mit Fahrzeugen, Tankungen und Ausgaben, bevor irgendetwas importiert wird.
- **Automatische Gruppierung identischer Fahrzeugdatensätze**, z. B. mehrfach gespeicherte Fahrzeuge mit gleichem Namen, Kennzeichen, Hersteller und Modell. Zugehörige Tankungen und Ausgaben bleiben erhalten.
- **Checkboxen** zur Auswahl der Fahrzeuge. Ohne Auswahl startet kein Import.
- Erkennung bereits importierter Tankungen/Ausgaben anhand ihrer ursprünglichen IDs, um wiederholte Imports nicht zu duplizieren.
- **Schreibgeschütztes Öffnen** der Quelldatei; Änderungen an der Zieldatenbank laufen in einer Transaktion.

**Achtung:** Übernommen werden derzeit Fahrzeuge, Tankungen und Ausgaben. **Nicht enthalten** sind Benutzerkonten, Passwörter, Freigaben, Anhänge/Rechnungsdateien, alte Quick Entries und Erinnerungen. Fahrerzuordnungen werden nicht übernommen. Der erfolgreiche Abschluss eines Imports auf einem echten Original-Datenbestand wurde noch nicht als getestet bestätigt.

[Ausführliche Migrationsanleitung](docs/MIGRATION-ORIGINAL-HAMMOND.md)

### 2. Wartungscenter und Dashboard
- Wartungen für **HU/TÜV, Inspektion und Service** mit Fälligkeitsdatum und/oder Kilometerstand.
- Übersicht über anstehende Wartungen direkt im Dashboard.
- Überarbeitete Kostenstatistiken mit Zeitraumfilter, Kostendiagrammen und Verbrauchsübersicht.
- Schneller Einstieg **„Tankung hinzufügen“** mit Fahrzeugauswahl.
- Die früheren **Quick Entries** wurden aus der Oberfläche entfernt.

### 3. Design, Mobilgeräte und Zahlenformatierung
- Bessere Fahrzeugdetails und Tabellen auf iPhone und Desktop.
- Dark-Mode-Korrekturen für Tabellen, Hover-Effekt und Auswahllisten.
- Währungen als formatierten Betrag, z. B. **76,63 €**, entsprechend den Einstellungen.
- Literangaben in Statistiken und Fahrzeugdetails auf **eine Nachkommastelle**, z. B. **42,7 Lt**, gerundet.
- Überarbeitete Import-Oberfläche mit besser lesbaren Hinweisen und Auswahlfeldern.

### 4. Noch experimentell
Die **Tankbeleg-Erkennung per Foto/OCR** ist weiterhin nicht zuverlässig. Bitte Tankdaten manuell prüfen und erfassen. Es findet keine automatische Buchung aus einem Foto statt.

Eine automatische Wartungsbenachrichtigung ist derzeit nicht enthalten.

## GitHub-Versionen

| Branch / Tag | Zweck |
| --- | --- |
| [`stable/hammond-modern`](https://github.com/Skoun181/hammond-modern/tree/stable/hammond-modern) | Standardbranch für aktuelle Stable-Installationen |
| [`develop/hammond-modern-next`](https://github.com/Skoun181/hammond-modern/tree/develop/hammond-modern-next) | Weiterentwicklung; Änderungen zunächst hier testen |
| [`v1.1.6`](https://github.com/Skoun181/hammond-modern/releases/tag/v1.1.6) | Stable-Release mit ZIP-Backup und Restore |
| [`v1.1.5`](https://github.com/Skoun181/hammond-modern/releases/tag/v1.1.5) | Frühere Version |
| [`v1.1.0`](https://github.com/Skoun181/hammond-modern/releases/tag/v1.1.0) | Frühere Version |
| [`v1.0.0`](https://github.com/Skoun181/hammond-modern/releases/tag/v1.0.0) | Erste Stable-Version |

Die Branches `stable` und `develop` können sich weiterentwickeln. Für einen **festen** Versionsstand das jeweilige Release-Tag auschecken. Es gibt keinen `master`-Branch mehr.

## Installation

### Voraussetzungen

Docker Engine (oder Synology Container Manager), AMD64/x86-64, ein freier Host-Port (Standard: 3006). Für Compose zusätzlich Docker Compose v2, für den Quellcode-Build Git. Die Shell-Beispiele setzen Linux und OpenSSL voraus. ARM64-Images werden derzeit nicht angeboten.

### Variante A – Docker Compose ohne Git (empfohlen)

```bash
mkdir -p hammond-modern/data/{config,assets}
cd hammond-modern
# Nur bei einer Neuinstallation ausführen:
(umask 077; printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)" > .env)
```

Datei `compose.yaml` erstellen:

```yaml
services:
  hammond-modern:
    image: skoun181/hammond-modern:1.1.6
    container_name: hammond-modern
    restart: unless-stopped
    ports:
      - "3006:3000"
    environment:
      TZ: Europe/Berlin
      JWT_SECRET: ${JWT_SECRET:?JWT_SECRET fehlt}
      TANKERKOENIG_API_KEY: ${TANKERKOENIG_API_KEY:-}
    volumes:
      - ./data/config:/config
      - ./data/assets:/assets
```

```bash
docker compose pull
docker compose up -d
docker compose ps
```

Browser: `http://SERVER-IP:3006`. Beim ersten Start Administrator einrichten.

### Variante B – Docker Pull und Docker Run (ohne Compose)

In einem neuen Installationsverzeichnis:

```bash
mkdir -p hammond-modern/data/{config,assets}
cd hammond-modern
# Nur bei einer Neuinstallation ausführen:
(umask 077; printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)" > .env)
printf 'TZ=Europe/Berlin\n' >> .env
docker pull skoun181/hammond-modern:1.1.6
docker run -d \
  --name hammond-modern \
  --restart unless-stopped \
  --env-file .env \
  -p 3006:3000 \
  -v "$PWD/data/config:/config" \
  -v "$PWD/data/assets:/assets" \
  skoun181/hammond-modern:1.1.6
```

Browser: `http://SERVER-IP:3006`. Docker Pull allein startet keinen Container.

### Variante C – Git Clone und lokaler Build

```bash
git clone --branch stable/hammond-modern https://github.com/Skoun181/hammond-modern.git
cd hammond-modern
# Optional: festen Release-Stand verwenden
git checkout v1.1.6
mkdir -p data/config data/assets
# Nur bei einer Neuinstallation ausführen:
(umask 077; printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)" > .env)
docker compose -f compose.test.yml up -d --build
docker compose -f compose.test.yml ps
```

Browser: `http://SERVER-IP:3006`. Ohne den optionalen Checkout wird der aktuelle Stable-Branch gebaut. Der technische Containername ist bei dieser Variante `hammond-modern-next`.

### Konfiguration

| Variable / Pfad | Bedeutung |
| --- | --- |
| `JWT_SECRET` | Pflicht: langer zufälliger Schlüssel; bei Updates beibehalten |
| `TANKERKOENIG_API_KEY` | Optional: eigene API-Zugangsdaten für deutsche Tankstellensuche |
| `TZ` | Zeitzone, Standard Europe/Berlin |
| `/config` | Persistente Konfiguration und SQLite-Datenbank |
| `/assets` | Persistente Anhänge und Rechnungen |

Optional in `.env` ergänzen: `TANKERKOENIG_API_KEY=DEIN_EIGENER_API_KEY`. Secrets niemals veröffentlichen. Die Verzeichnisse und die .env-Datei sichern.

### Host-Port ändern

Das Format lautet `HOSTPORT:CONTAINERPORT`. Intern bleibt Hammond auf **3000**. Beispiel: Host-Port 8085.

In `compose.yaml` bzw. `compose.test.yml`:

```yaml
ports:
  - "8085:3000"
```

Danach `docker compose up -d` (lokaler Build: `docker compose -f compose.test.yml up -d`). Bei Docker Run stattdessen `-p 8085:3000` verwenden; einen vorhandenen Container hierfür nach Sicherung und Prüfung seiner Bind Mounts stoppen, entfernen und mit demselben vollständigen Run-Befehl neu erstellen. Zugriff: `http://SERVER-IP:8085`.

Für einen Reverse Proxy auf demselben Host kann die Bindung auf `127.0.0.1:8085:3000` eingeschränkt werden. Einen HTTP-Port nicht ungeschützt ins Internet freigeben; HTTPS oder VPN verwenden.

### Synology / Dockhand

Einen normalen Compose-Stack erstellen und Variante A verwenden. Relative Datenpfade bei Bedarf durch absolute DSM-Pfade ersetzen. JWT_SECRET und optional TANKERKOENIG_API_KEY in Dockhand als Environment-Variablen setzen. Für eine eigene LAN-IP per macvlan gibt es die [Synology-Compose-Vorlage](docker-compose.dockerhub.yml) und die [DSM-Anleitung](docs/SYNOLOGY.md). Netzwerkname, freie IP und Volume-Pfade anpassen. Bei macvlan ohne Port-Mapping ist die Anwendung direkt auf Container-IP:3000 erreichbar.

### Updates und Backups

Vor Updates den Container stoppen und die persistenten Verzeichnisse sowie .env sicher sichern (alternativ konsistente SQLite-Online-Sicherung plus Anhänge). Keine Datenverzeichnisse löschen. Den JWT_SECRET beibehalten.

- Compose: Image-Tag auf die gewünschte veröffentlichte Version ändern, dann `docker compose pull && docker compose up -d`.
- Docker Run: neues Image pullen, alten Container stoppen und entfernen, vollständigen Run-Befehl mit neuem Tag und unveränderten Bind Mounts erneut ausführen.
- Git-Build: `git fetch origin --tags`, gewünschten Release-Tag auschecken oder `git switch stable/hammond-modern` und `git pull --ff-only origin stable/hammond-modern`; dann `docker compose -f compose.test.yml up -d --build`.

Der feste Tag `1.1.6` ist für reproduzierbare Installationen empfohlen. `latest` aktualisiert laufende Container nicht automatisch. Stable und Develop nicht gegen dieselben Datenverzeichnisse betreiben.

## Backup, Wiederherstellung und Serverumzug

> **Ab Stable v1.1.6 verfügbar.** Export und ZIP-Import wurden praktisch getestet. Vor einem produktiven Umzug immer extern sichern.

Unter **Backup & Umzug** können Administratoren jetzt eine **portable ZIP-Sicherung**
erstellen, das ZIP auf einer frischen Installation prüfen und alle bisherigen
Anwendungsdaten wiederherstellen.

- **Im ZIP:** Fahrzeuge, Tankungen, Ausgaben, Wartungen, Fahrer, Zuordnungen,
  Einstellungen und Uploads/Rechnungen aus `/assets`, dazu ein Manifest mit
  Exportzeit, App-/Schema-Version, Anzahl der Datensätze und SHA-256-Prüfsummen.
- **Nicht im ZIP:** Passwörter und Passwort-Hashes, JWT-Secret, API-Keys, Tokens,
  `.env`, Docker-Compose- oder Netzwerkkonfiguration.
- **Sicherheit:** Admin-only, Größenlimits, Integritäts- und Schema-Prüfung,
  sichere ZIP-Pfade, Vorab-Sicherheitsbackup und SQLite-Transaktion.
- **Wichtig:** Der Import **ersetzt** vorhandene Daten. Für den neuen Server
  muss ein Admin mit **gleicher E-Mail-Adresse** wie im Backup angelegt werden;
  dessen Passwort auf dem Zielserver bleibt erhalten. Weitere Konten benötigen
  neu vergebene Passwörter (Admin-Benutzerverwaltung).
- **Recovery:** Ein dauerhaftes Journal ermöglicht beim Neustart die automatische
  Wiederherstellung der ursprünglichen Daten bei einem unterbrochenen Import.
  Schäden an Datenträgern erfordern dennoch ein externes Offline-Backup.
  Jeden Restore zuerst in einer Testinstanz üben.

### Frische Installation mit benannten Volumes (optional)

```bash
git clone --branch v1.1.6 https://github.com/Skoun181/hammond-modern.git
cd hammond-modern
(umask 077; printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)" > .env)
docker compose up -d --build
```

Die neue `docker-compose.yml` verwendet benannte Volumes
`hammond_config` und `hammond_assets`. Der öffentliche Port ist über
`HAMMOND_PORT` in der `.env` einstellbar (Standard `3006`, Containerport `3000`).

**Bestehende Installationen:** Bisherige Compose-Datei, Projektnamen, Mounts und JWT_SECRET beim Update beibehalten. Nicht einfach auf neue leere Volumes wechseln. Alte v1.1.5-Instanz zuerst auf v1.1.6 aktualisieren und dann exportieren.

### Umzug in fünf Schritten

1. Als Admin auf dem alten Server **Backup & Umzug → Backup herunterladen (.zip)** wählen und extern sichern.
2. Auf dem neuen Server v1.1.6 installieren und einen Admin mit derselben E-Mail-Adresse wie im Backup anlegen. Neues Passwort und JWT_SECRET sind möglich.
3. ZIP auswählen und **Backup prüfen**. Maximal 512 MiB ZIP und 1 GiB entpackte Daten.
4. Warnung bestätigen, Checkbox setzen und `ERSETZEN` eingeben. **Daten endgültig wiederherstellen** startet den Import mit Sicherheitsbackup.
5. Mit Ziel-Admin-Passwort neu anmelden, weitere Benutzerpasswörter setzen und Fahrzeuge, Tankungen, Wartungen und Rechnungen prüfen. Alten Server erst danach abschalten.

**[Vollständige Schritt-für-Schritt-Anleitung für Anfänger: Export, Import,
Serverumzug, Datensicherheit und Offline-Volume-Backups](docs/BACKUP-RESTORE-SERVERUMZUG.md)**

## Original-Hammond zu Hammond Modern migrieren

> **Bitte vor der Migration zwei unabhängige Sicherungen anlegen:** eine vom ursprünglichen Hammond und eine vom bestehenden Hammond Modern. Niemals die alte `hammond.db` direkt in das neue Konfigurationsverzeichnis kopieren.

1. **Original-Hammond stoppen** oder eine konsistente SQLite-Online-Sicherung anlegen. Bei laufendem Betrieb können ungeschriebene Daten in `hammond.db-wal` liegen.
2. Die ursprüngliche `hammond.db` auf deinen Computer kopieren.
3. In Hammond Modern **Datenimport → Hammond-Migration starten** öffnen.
4. Die Datenbank über **„Datenbank auswählen“** hochladen und **„Sicherung prüfen“** klicken.
5. Die **Vorschau** prüfen. Identische Quellfahrzeuge werden zusammengefasst; die angezeigten Tankungen und Ausgaben prüfen.
6. Nur die gewünschten Fahrzeuge **per Checkbox** auswählen.
7. Erst nach eigener Kontrolle den Import bestätigen. Anschließend Fahrzeugdaten und Summen stichprobenartig überprüfen.

Doppelte Fahrzeuge werden nur bei passender Identität gruppiert; unterschiedliche Angaben können zu getrennten Fahrzeugen führen. Beim Import werden keine Währungen oder Kilometer/Meilen umgerechnet. Anhänge müssen vorerst separat behandelt werden.

[Mehr zur Migration, Grenzen und Fehlerbehebung](docs/MIGRATION-ORIGINAL-HAMMOND.md)

## Bedienung im Alltag

- **Fahrzeug hinzufügen:** Fahrzeug mit Name, Kennzeichen, Kraftstoffart und weiteren Angaben anlegen.
- **Tankung hinzufügen:** Fahrzeug auswählen; Datum, Kilometerstand, Liter, Preis und Tankstelle eingeben; speichern. Die Fotoerkennung ist optional und experimentell.
- **Ausgabe hinzufügen:** Werkstatt- oder sonstige Kosten samt optionaler Rechnung erfassen.
- **Wartungscenter:** z. B. „TÜV“ oder „Ölwechsel“ mit Datum bzw. Kilometerstand anlegen und später erledigt markieren.
- **Dashboard:** Kostenzeitraum auswählen und Ausgaben sowie anstehende Wartungen ansehen.
- **Design wechseln:** Dark-/Light-Mode über die Navigation umschalten.

## Fehlerdiagnose

```bash
docker compose -f compose.test.yml ps
docker compose -f compose.test.yml logs --tail=100
docker compose -f compose.test.yml restart
```

Beim **Original-Hammond-Import** nur eine echte SQLite-`hammond.db`-Datei verwenden (max. 50 MB). Für **Backup & Umzug** dagegen ausschließlich eine portable ZIP-Sicherung (max. 512 MiB). Falls die Prüfung scheitert, zuerst prüfen, ob die Originaldatenbank konsistent gesichert wurde.

Für den Fernzugriff **HTTPS über Reverse Proxy oder VPN** verwenden und den HTTP-Port nicht ungeschützt im Internet veröffentlichen.

## Sicherheit und Einschränkungen

- Hammond Modern enthält ältere Abhängigkeiten aus dem Originalprojekt; ein vollständiges Sicherheits-Audit fehlt.
- Keine garantierte Kompatibilität mit allen sehr alten oder individuell veränderten Hammond-Datenbanken.
- Die Migrations-*Vorschau* einschließlich Duplikatgruppierung wurde praktisch bestätigt. Der portable ZIP-Restore wurde praktisch getestet; die vollständige Original-Hammond-Migration und alle Mehrnutzer-/Anhänge-Sonderfälle sind weiterhin nicht umfassend validiert.
- Die Tankbeleg-OCR-Erkennung bleibt experimentell.
- Oberfläche und Dokumentation sind teilweise noch nicht vollständig übersetzt.
- Der Synology-/Dockhand-Start des Docker-Hub-Images wurde bestätigt; das ersetzt keinen vollständigen Funktions- oder Sicherheitstest auf DSM.

## Screenshots

### Dashboard
![Hammond Modern – Dashboard](docs/screenshots/dashboard.png)

### Tankung erfassen
![Hammond Modern – Tankung erfassen](docs/screenshots/tankung.png)

### Tankstellensuche
![Hammond Modern – Tankstellensuche](docs/screenshots/tankstellensuche.png)

## Weitere Dokumentation

- [Migration vom Original-Hammond](docs/MIGRATION-ORIGINAL-HAMMOND.md)
- [Versionshistorie](CHANGELOG-MODERN.md)
- [Testnotizen: Fahrzeug- und Tankstellensuche](docs/FAHRZEUG-TANKSTELLEN-TEST.md)
- [Weitere Funktionen](docs/HAMMOND-MODERN-NEUE-FUNKTIONEN.md)
- [Synology-Installation](docs/SYNOLOGY.md)
- [Backup, Wiederherstellung und Serverumzug](docs/BACKUP-RESTORE-SERVERUMZUG.md)

## Herkunft und Lizenz

Hammond Modern basiert auf [akhilrex/hammond](https://github.com/akhilrex/hammond) von Akhil Gupta. Vielen Dank an den ursprünglichen Autor und die bisherigen Mitwirkenden. Das Repository steht unter **[AGPL-3.0](LICENSE)**; ursprüngliche Lizenz- und Copyright-Hinweise bleiben erhalten.

[Repository](https://github.com/Skoun181/hammond-modern) · [Issues](https://github.com/Skoun181/hammond-modern/issues) · [Releases](https://github.com/Skoun181/hammond-modern/releases)
