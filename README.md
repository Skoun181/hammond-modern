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

## Installation: Debian-LXC / Docker

Die folgenden Befehle entsprechen der bisher getesteten Konfiguration (Linux AMD64 / x86_64; Docker Compose). Für einen öffentlich erreichbaren Betrieb sind zusätzliche Schutzmaßnahmen nötig.

### 1. Repository klonen

```bash
git clone --branch stable/hammond-modern git@github.com:Skoun181/hammond-modern.git
cd hammond-modern
```

Alternativ ohne SSH-Schlüssel:

```bash
git clone --branch stable/hammond-modern https://github.com/Skoun181/hammond-modern.git
cd hammond-modern
```

### 2. Zugangsdaten lokal konfigurieren

Erstelle im Projektverzeichnis die Datei `.env`:

```bash
printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)" > .env
chmod 600 .env
```

Für die optionale Tankstellensuche zusätzlich **deinen eigenen** Tankerkönig-Key in `.env` eintragen:

```dotenv
TANKERKOENIG_API_KEY=DEIN_EIGENER_API_KEY
```

Den vorhandenen `JWT_SECRET` bei späteren Updates **nicht neu generieren**. Weder `.env`, API-Keys noch Zugangsdaten in Git einchecken.

### 3. Anwendung bauen und starten

```bash
docker compose -f compose.test.yml up -d --build
docker compose -f compose.test.yml ps
```

Die Anwendung ist danach im lokalen Netzwerk unter **`http://SERVER-IP:3006`** erreichbar.

Beim ersten Start in der Weboberfläche einen Administrator anlegen und die persönlichen Einheiten/Währung konfigurieren.

### 4. Aktualisieren

Für einen Stable-Checkout:

```bash
git pull --ff-only origin stable/hammond-modern
docker compose -f compose.test.yml up -d --build
```

Für einen bestehenden Entwicklungs-Checkout entsprechend `develop/hammond-modern-next` verwenden.

**Niemals bei einem normalen Update `data/` löschen:** Dort liegen Datenbank und Anhänge.

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
