# Hammond Modern – erster Implementierungsstand

## Umgesetzt
- Dark/Light-Umschalter in der Navigation, Einstellung im Browser gespeichert
- Zentrale Theme-Variablen und modernisierte Karten, Formulare und Tabellen
- Deutsches Währungsformat im Statistik-Widget (Werte/Berechnungen unverändert)
- Synology-Compose-Datei für AMD64 und Port 3005

## Noch offen
- Vollständige Übersetzung und Formatierung aller Fahrzeug- und Ausgabenseiten
- Sidebar und umfassende Dashboard-Neugestaltung
- Update der veralteten Vue-2-/Node-14-/Go-1.16-Buildkette
- Docker-Build, Integrationstests und Sicherheitsprüfung

## Start (nach erfolgreichem Build)
1. `.env.example` nach `.env` kopieren und `JWT_SECRET` durch ein langes zufälliges Secret ersetzen.
2. Pfade in `docker-compose.synology.yml` an NAS anpassen.
3. `docker compose -f docker-compose.synology.yml up -d --build` ausführen.

**Wichtig:** Build noch nicht verifiziert. Der Original-Dockerfile verwendet alte Build-Images.
