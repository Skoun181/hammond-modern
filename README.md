# Hammond Modern 🚗

**Modernisierter Self-Hosted-Fahrzeugmanager auf Basis von [akhilrex/hammond](https://github.com/akhilrex/hammond).**

Hammond Modern ist ein Community-Fork von Hammond mit modernisierter Oberfläche, deutschen Beschriftungen und zusätzlichen Funktionen rund um Fahrzeuge, Tankungen, Fahrer und Werkstattkosten. Das Projekt ist **nicht offiziell mit dem Originalautor verbunden**.

> **Projektstand: 17. November 2026.** Die aktuelle Version wurde auf einer privaten Debian-13-LXC-Testinstanz mit Docker erfolgreich auf ihre Kernfunktionen geprüft. Das ist kein vollständiger Sicherheits- oder Produktionstest.

## Funktionen

- **Fahrzeuge:** Mehrere Fahrzeuge verwalten, anlegen und bearbeiten; Automarken mit Suchfunktion sowie Modellvorschläge über NHTSA vPIC. Modelle lassen sich bei Bedarf manuell eingeben.
- **Tankungen:** Kraftstoffmenge, Preis, Kilometerstand, Tankstelle und optionalen Fahrer erfassen.
- **Tankstellensuche:** Deutsche Postleitzahl eingeben und Stationen über Nominatim und Tankerkönig suchen (eigener API-Key erforderlich).
- **Fahrer:** Fahrerprofile ohne separates Benutzerkonto verwalten und Tankungen zuordnen.
- **Werkstatt und Ausgaben:** Werkstatt, durchgeführte Arbeiten, sonstige Kosten und Rechnungsanhänge erfassen.
- **Dashboard:** Fahrzeugübersicht, Tankmengen für den aktuellen Monat/das Jahr, Kostenübersicht und Verbrauchsberechnung aus passenden Volltankungen.
- **Oberfläche:** Modernisiertes Layout mit Dark-/Light-Mode und teilweise deutscher Benutzeroberfläche.
- **Hammond-Grundfunktionen:** Mehrbenutzer-Unterstützung, Fahrzeugfreigaben, Belege und weitere Funktionen des ursprünglichen Projekts, soweit im Fork erhalten.

Die Verbrauchsberechnung benötigt mindestens zwei geeignete Volltankungen mit steigenden Kilometerständen. Die NHTSA-Modellsuche ist besonders auf US-Daten ausgerichtet; europäische Modelle können fehlen.

## Entwicklungsstand und Branches

| Branch | Bedeutung |
| --- | --- |
| [`master`](https://github.com/Skoun181/hammond-modern/tree/master) | Ursprüngliche Hammond-Codebasis als Referenz; keine Hammond-Modern-Installation |
| [`stable/hammond-modern`](https://github.com/Skoun181/hammond-modern/tree/stable/hammond-modern) | Getesteter Stand der Hammond-Modern-Erweiterungen |
| [`develop/hammond-modern-next`](https://github.com/Skoun181/hammond-modern/tree/develop/hammond-modern-next) | Laufende Weiterentwicklung |

**Wichtig:** Der GitHub-Standardbranch ist derzeit `master`. Für Hammond Modern beim Klonen oder nach dem Klonen **`stable/hammond-modern`** oder zum Entwickeln **`develop/hammond-modern-next`** auswählen. Der Stable-Branch ist keine automatische Docker-Image-Veröffentlichung.

## Screenshots der neuen Oberfläche

Hier sind zwei Plätze für echte Screenshots von Hammond Modern reserviert. Sobald Bilder aus der laufenden Instanz vorliegen, können sie unter den angegebenen Namen in `docs/screenshots/` abgelegt und die auskommentierten Bildzeilen aktiviert werden.

| Dashboard und Fahrzeugübersicht | Tankung / Tankstellensuche |
| --- | --- |
| 📸 **Screenshot 1 – Dashboard**<br>Geplanter Pfad: `docs/screenshots/dashboard.png` | 📸 **Screenshot 2 – Tankstellen oder Tankung**<br>Geplanter Pfad: `docs/screenshots/tankstellen.png` |

<!-- Nach dem Hochladen der echten Bilder diese Zeilen einkommentieren:
![Hammond Modern – Dashboard](docs/screenshots/dashboard.png)
![Hammond Modern – Tankstellen- und Tankungsansicht](docs/screenshots/tankstellen.png)
-->

**Tipp:** Vor Veröffentlichung persönliche Kennzeichen, Standortdaten und Tankbelege in den Screenshots unkenntlich machen.

## Installation – Docker und Docker Compose

Die Anleitung orientiert sich an der Installation des ursprünglichen Hammond, verwendet aber **unseren Quellcode**. Ein fertig veröffentlichtes Docker-Hub-Image wird hier nicht vorausgesetzt. Der bisher getestete Weg ist **Docker Compose mit lokalem Build** auf Linux AMD64 (x86_64), unter anderem auf Debian 13 im LXC.

### Voraussetzungen

- Linux AMD64 / x86_64 mit Docker Engine und `docker compose` (Compose-Plugin)
- Git zum Herunterladen des Repositories
- Freier Host-Port `3006` (oder Portbelegung anpassen)
- Für die Tankstellensuche optional ein eigener Tankerkönig-API-Key

Beim Betrieb in einem Proxmox-LXC muss Docker bereits funktionsfähig eingerichtet sein. Diese Anleitung verändert keine Proxmox-Container-Einstellungen.

### 1. Repository herunterladen

**Empfohlen: Stable-Version per HTTPS**

```bash
git clone --branch stable/hammond-modern https://github.com/Skoun181/hammond-modern.git
cd hammond-modern
```

Alternativ per SSH, wenn ein GitHub-SSH-Schlüssel eingerichtet ist:

```bash
git clone --branch stable/hammond-modern git@github.com:Skoun181/hammond-modern.git
cd hammond-modern
```

Für die Weiterentwicklung kann stattdessen der Branch `develop/hammond-modern-next` verwendet werden. `master` enthält die ursprüngliche Hammond-Codebasis und ist **nicht** die hier beschriebene Modern-Version.

### 2. Konfiguration (`.env`)

Im Projektverzeichnis ein dauerhaftes JWT-Secret erstellen:

```bash
printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)" > .env
chmod 600 .env
```

**Optional:** In derselben `.env`-Datei die Tankstellensuche aktivieren:

```dotenv
TANKERKOENIG_API_KEY=DEIN_EIGENER_API_KEY
```

Den Wert von `JWT_SECRET` bei späteren Neustarts oder Updates **nicht ändern**. Die `.env`-Datei nicht veröffentlichen oder zu Git hinzufügen.

### 3A. Empfohlen: Docker Compose

Die mitgelieferte [`compose.test.yml`](compose.test.yml) baut das Image aus diesem Repository und bindet zwei lokale Datenverzeichnisse ein:

```bash
docker compose -f compose.test.yml up -d --build
docker compose -f compose.test.yml ps
```

Anschließend im Browser aufrufen:

```text
http://SERVER-IP:3006
```

Beim ersten Aufruf Administrator anlegen und gewünschte Einheiten sowie Währung einstellen.

**Auszug aus unserer Compose-Konfiguration:**

```yaml
services:
  hammond-modern:
    build: .
    image: hammond-modern:next
    container_name: hammond-modern-next
    ports:
      - "3006:3000"
    environment:
      TZ: Europe/Berlin
      JWT_SECRET: ${JWT_SECRET:?JWT_SECRET fehlt}
      TANKERKOENIG_API_KEY: ${TANKERKOENIG_API_KEY:-}
    volumes:
      - ./data/config:/config
      - ./data/assets:/assets
    restart: unless-stopped
```

Die vollständige, maßgebliche Datei liegt im Repository. `data/config` enthält die Konfiguration und SQLite-Datenbank, `data/assets` die Anhänge.

### 3B. Alternativ: Docker Run (wie beim Original)

Wer **kein Compose** verwenden möchte, kann dasselbe lokale Image manuell starten. Dazu ebenfalls zunächst das Repository klonen, die `.env` anlegen und das Image bauen:

```bash
docker build -t hammond-modern:local .
mkdir -p data/config data/assets
```

Anschließend starten:

```bash
docker run -d \
  --name hammond-modern \
  --restart unless-stopped \
  -p 3006:3000 \
  --env-file .env \
  -e TZ=Europe/Berlin \
  -v "$PWD/data/config:/config" \
  -v "$PWD/data/assets:/assets" \
  hammond-modern:local
```

Auch hier lautet die Adresse `http://SERVER-IP:3006`. Bei `docker run` werden beide Variablen aus `.env` geladen, sofern eingetragen. **Nur eine** der Varianten 3A oder 3B gleichzeitig mit Host-Port `3006` verwenden.

**Wichtig:** Nicht `docker run akhilrex/hammond` verwenden; das startet das originale Hammond und nicht unseren Fork.

### 4. Aktualisieren

Für die Compose-Installation auf dem Stable-Branch:

```bash
git switch stable/hammond-modern
git pull --ff-only origin stable/hammond-modern
docker compose -f compose.test.yml up -d --build
```

Für die manuelle Docker-Run-Variante nach dem `git pull` das Image erneut mit `docker build -t hammond-modern:local .` bauen und den bestehenden Container anschließend mit denselben `docker run`-Parametern neu erstellen. Die gemounteten Datenverzeichnisse dabei erhalten.

**Bei normalen Updates `data/config` und `data/assets` nicht löschen.**

### 5. Betrieb und Fehleranalyse

```bash
# Compose-Container und Status anzeigen
docker compose -f compose.test.yml ps

# Letzte Logzeilen
docker compose -f compose.test.yml logs --tail=100

# Container neu starten
docker compose -f compose.test.yml restart

# Container stoppen/entfernen – Datenverzeichnisse bleiben bestehen
docker compose -f compose.test.yml down
```

Bei der manuellen Variante stattdessen `docker logs hammond-modern` oder `docker restart hammond-modern` verwenden.

Eine direkte Freigabe des HTTP-Ports im Internet ist nicht empfohlen. Für externen Zugriff HTTPS und einen abgesicherten Reverse Proxy oder VPN verwenden.

## Docker-Konfiguration

Die vorhandene Datei [`compose.test.yml`](compose.test.yml) enthält:

| Einstellung | Wert |
| --- | --- |
| Containername | `hammond-modern-next` |
| Build | Aus dem lokalen `Dockerfile` |
| Host-Port | `3006` → Container-Port `3000` |
| Zeitzone | `Europe/Berlin` |
| Konfiguration / SQLite | `./data/config` → `/config` |
| Anhänge / Rechnungen | `./data/assets` → `/assets` |
| Pflichtvariable | `JWT_SECRET` |
| Optionale Variable | `TANKERKOENIG_API_KEY` |

Der Name `hammond-modern-next` stammt aus der bestehenden Test-Compose-Datei und gilt auch dann, wenn der Stable-Branch verwendet wird.

Nützliche Befehle:

```bash
docker compose -f compose.test.yml logs --tail=100
docker compose -f compose.test.yml restart
docker compose -f compose.test.yml down
```

`down` entfernt den Container, **nicht** die Bind-Mount-Datenverzeichnisse. Das Löschen von `data/config` oder `data/assets` entfernt dagegen persistente Daten.

## Technologie

- Backend: **Go**, **Gin**, **GORM**, **SQLite**
- Frontend: **Vue 2**, **Vuex**, **Buefy**
- Build: mehrstufiges Dockerfile mit **Go 1.25**, **Node.js 22** und Debian Bookworm
- Laufzeit: Linux AMD64 / x86_64

Teile der Frontend-Abhängigkeiten und des Authentifizierungs-Stacks stammen aus der älteren Hammond-Codebasis.

## Getestete Funktionen

Auf der privaten Testinstanz wurden folgende Abläufe erfolgreich überprüft:

1. Fahrzeug anlegen, bearbeiten und speichern
2. Automarken und passende Modelle suchen
3. Tankstellen per PLZ suchen und übernehmen
4. Tankung mit Fahrer speichern
5. Tankkosten und Verbrauch im Dashboard anzeigen
6. Werkstattausgabe mit Rechnung erfassen
7. Seite neu laden und gespeicherte Daten wiederfinden

Diese Tests bestätigen die genannten Anwendungsabläufe, ersetzen aber weder automatisierte Regressionstests noch eine Sicherheitsprüfung.

## Einschränkungen und Sicherheit

- Derzeit **nicht** direkt ins öffentliche Internet stellen. Für Fernzugriff HTTPS, Reverse Proxy/VPN und Zugriffsschutz nutzen.
- `/config` enthält die SQLite-Datenbank, `/assets` gespeicherte Anhänge; beides kann sensible personenbezogene Daten enthalten.
- Eine sinnvolle Backup-Strategie sollte vor produktiver Nutzung eingerichtet werden.
- Die Tankstellensuche hängt von externen Diensten sowie deren Nutzungsbedingungen und Verfügbarkeit ab.
- Die Oberfläche ist noch nicht durchgängig übersetzt.
- Synology-/DSM-Bereitstellung ist dokumentiert, für den aktuellen Stable-Stand jedoch **nicht als getestet bestätigt**.

## Weitere Dokumentation

- [Fahrzeug- und Tankstellensuche testen](docs/FAHRZEUG-TANKSTELLEN-TEST.md)
- [Erweiterte Funktionen](docs/HAMMOND-MODERN-NEUE-FUNKTIONEN.md)
- [Synology-Installationshinweise (separat prüfen)](docs/SYNOLOGY.md)
- [Änderungsverlauf](CHANGELOG-MODERN.md)

Ältere Notizen oder Testbuild-Anleitungen können abweichende Ports oder überholte Entwicklungsstände enthalten. Für den hier dokumentierten Compose-Start gilt **Port 3006**.

## Ursprung und Lizenz

Hammond Modern basiert auf dem Open-Source-Projekt **[akhilrex/hammond](https://github.com/akhilrex/hammond)** von Akhil Gupta. Danke an den Originalautor und die bisherigen Mitwirkenden.

Das Repository enthält die **[GNU Affero General Public License v3.0 (AGPL-3.0)](LICENSE)**. Originale Lizenz- und Copyright-Hinweise bleiben erhalten.

**Repository:** [Skoun181/hammond-modern](https://github.com/Skoun181/hammond-modern) · [Issues](https://github.com/Skoun181/hammond-modern/issues)
