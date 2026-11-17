# Hammond Modern – Entwicklungsstand November 2026

Dies ist ein **ungetesteter Entwicklungs-Build** für deinen Debian-13-LXC, noch keine produktionsreife Version.

## Erneut bauen (Port 3005)

1. Die bisherige Datenbank sichern (ohne laufenden Schreibzugriff):

```bash
docker compose -f compose.test.yml down
mkdir -p ../hammond-backup
cp -a data/config ../hammond-backup/config-$(date +%Y%m%d-%H%M%S)
cp -a data/assets ../hammond-backup/assets-$(date +%Y%m%d-%H%M%S)
```

2. Neues Projekt in einem **separaten Ordner** entpacken. Die bisherigen `data/`-Ordner nicht löschen. Vorhandene `.env` sicher übertragen, **nicht** in Git hochladen.

3. Im neuen Ordner `compose.test.yml` nutzen, `.env` mit `JWT_SECRET` erstellen (falls noch nicht vorhanden). Für die Tankstellensuche optional `TANKERKOENIG_API_KEY=...` eintragen. Die Variable ist serverseitig.

4. Bestehende Sicherungen vorsichtig in `data/config/` und `data/assets/` **nur in einer Testkopie** verwenden. DB-Migrationen passieren automatisch beim Start.

5. Build und Logs:

```bash
set -o pipefail
docker compose -f compose.test.yml build --progress=plain 2>&1 | tee build.log
docker compose -f compose.test.yml up -d
docker compose -f compose.test.yml ps
docker compose -f compose.test.yml logs --tail=100
```

Danach `http://LXC-IP:3005` öffnen. Unbedingt Fahrer, Tankung mit Fahrer, Rechnung speichern/herunterladen, Fahrzeugmarke/Modell, Tankstelle und Dashboard testen.

## Einschränkungen

- Der Quellcode konnte in der hier verwendeten Arbeitsumgebung nicht vollständig mit Docker gebaut werden, weil die externen Module fehlen. Die Go- und Vue-Dateien sind daher nicht Ende-zu-Ende-geprüft.
- NHTSA-Modellliste ist nicht vollständig für Europa; manuelle Eingabe bleibt möglich.
- Tankerkönig braucht einen eigenen API-Key; Abfragen nur nach explizitem Klick, maximal ca. einmal/minut für neue PLZ und 15 Minuten Cache. Bei nicht unterstützter Nutzung keinen Key eintragen.
- Übersetzung der bestehenden Oberfläche ist nur teilweise fertig. Ältere Beschriftungen können noch Englisch enthalten.
- Fahrer-/Werkstattdaten und hochgeladene Rechnungen benötigen geschützten Zugriff (VPN/HTTPS). Alt-Code hat ältere Abhängigkeiten und muss vor produktiver Bereitstellung sicherheitsgeprüft werden.
