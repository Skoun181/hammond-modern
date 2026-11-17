# Hammond Modern – Erweiterung (Entwicklungsstand)

## Vor dem ersten Start

**Backup der vorhandenen `/config/hammond.db` und `/assets` erstellen.** `AutoMigrate` ergänzt Tabellen/Spalten; einen automatischen Rollback gibt es nicht.

## Neue Funktionen

- Fahrer ohne eigenes Login mit Name, optionalem Führerschein-Ablaufdatum und Notiz; Bearbeiten/Löschen. Zugriff ist nach Benutzer eingeschränkt. Kein Ablauf-Reminder oder Führerscheinnummer-Speicher.
- Tankung mit optionalem Fahrer; `userId` bleibt der Ersteller des Eintrags.
- Werkstattausgabe mit Werkstatt, Arbeiten und Rechnung (PDF/JPG/PNG bis 10 MB). Rechnungen in `/assets`, UUID in der Datenbank. Andere Ausgaben bleiben möglich.
- Automarken: 110 offline verfügbare Markennamen mit Suche, Modellvorschläge online aus NHTSA vPIC; europäische Modelle können fehlen und sind manuell eingebbar.
- Tankstellen: Suche nach deutscher 5-stelliger PLZ auf ausdrücklichen Klick. Nominatim zur Geocodierung, Tankerkönig zur Umkreissuche. Es wird ein eigener API-Key benötigt. API-Bedingungen (insbesondere Rate Limits und Nutzungsberechtigung) beachten. Ergebnisse 15 Minuten zwischengespeichert. Standorte werden nicht automatisch im Hintergrund aktualisiert; die nächste Suche holt Daten neu.
- Dashboard je Fahrzeug: getankte Liter im laufenden Kalendermonat und Kalenderjahr; Verbrauch gewichtet zwischen zwei Volltankungen, einschließlich eventueller Zwischentankungen. Bei fehlender Tankung oder unplausiblen Odometerständen wird kein Wert aus dieser Teilstrecke gerechnet.
- Kostenübersicht: Die bisherige `/api/me/stats`-API summiert bereits über alle für den angemeldeten Benutzer zugänglichen Fahrzeuge (nicht über fremde Benutzerkonten).

## Installation / Test

1. Bestehende Daten sichern.
2. Quellcode ins Projektverzeichnis entpacken.
3. `.env` mit `JWT_SECRET` belassen. Für Tankstellen optional `TANKERKOENIG_API_KEY=...` ergänzen (Key niemals ins Git-Repository legen).
4. `docker compose -f compose.test.yml up -d --build` (nutzt Port 3005).
5. In neuem Testsystem zunächst Benutzer, Fahrer, Fahrzeug, Tankung, Ausgabe und Rechnung prüfen.

## Einschränkungen

- Ungetestet in dieser Umgebung: Docker-Build und Browser-End-to-End-Tests. Auf dem Test-LXC bauen und Fehlerlogs prüfen.
- Keine vollständige Übersetzung aller historischen Vue-Komponenten. Hauptformulare und Navigation sind teilweise eingedeutscht.
- NHTSA/vPIC ist eine US-Datenquelle und keine vollständige deutsche Typgenehmigungsdatenbank.
- Führerscheingültigkeit ist ein Erinnerungsdatum, keine automatische Prüfung der Fahrerlaubnis.
- Datenschutz: Rechnungen und Fahrerangaben sind persönliche Daten; nur mit TLS/VPN zugänglich machen, Zugriffsrechte prüfen und Backups schützen.
