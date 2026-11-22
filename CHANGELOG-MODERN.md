# Hammond Modern – Änderungsverlauf

## 1.1.10-dev – Noch nicht veröffentlicht

- Centgenaue Restschuldmodellierung mit optionalem monatlichen Berechnungszins und abweichendem Zinsanteil der ersten Rate.

- Anzahlung mit einmaliger Ausgabenerfassung, Vertragszinsen und Nettodarlehen für neue und bestehende Finanzierungen; Gesamtaufwand und modellierte Restschuld ohne Änderung des Ratenplans.

## v1.1.9 – Wartungserinnerungen, Fixkosten und Finanzierung (22. November 2026)

- Bestehende Finanzierungen ab ursprünglicher erster Fälligkeit erfassen; historische Raten nachtragen oder ausdrücklich bereits erfasste Raten berücksichtigen, mit Vorschau und Vertragsende.

- Intelligente Wartungserinnerungen nach Datum/km und einstellbarer Vorwarnung im Dashboard und optionalem ntfy-Push; Folgetermine und Historie bei wiederkehrenden Wartungen.
- Fixkosten und Finanzierung: automatische Sollbuchung, feste Intervalle und Ratenzahl, optionale Schlussrate, Fortschritt und ausstehende Ratensumme; transaktionale Buchung ohne doppelte Erzeugung.
- Neue Planungsdaten in portablen ZIP-Backups; kompatible alte Backups ohne neue Tabellen bleiben importierbar.

## v1.1.8 – Automatische Reifenlaufleistung und E5/E10

- Reifenlaufleistung auf Basis des Montage-Kilometerstands und späterer km-Tankungen; Abschluss beim Abmontieren, Fortsetzung beim Wiederaufziehen.
- Gemeinsamer Reifenwechsel mit Kilometerstand und Radpositionen; automatische Laufleistung in ZIP-Backups, kompatible ältere Backups ohne Montageintervalle.
- Petrol → Super E5, zusätzlicher Kraftstofftyp Super E10 ohne Änderung der bestehenden Enum-IDs.

## v1.1.7 – Reifenverwaltung und Fahrzeugübersicht (21. November 2026)

- Mehrfachanlage von 1–4 Reifen mit separaten Datensätzen, eindeutigen montierten Radpositionen und vollständigem Rollback bei Konflikten.
- Neue kontrastreiche Reifenkacheln mit großem Fahrzeugnamen und Kennzeichen; Fahrzeugkacheln mit Kilometerstand, Historie, Reifenzustand und montierten Größen.
- Überarbeiteter Develop-Stand praktisch getestet (Anwenderbestätigung); automatisierte Backend-Tests, Frontend-Build und GHCR-Starttest erfolgreich.
- Reifen-Inventar pro Nutzer: Fahrzeugzuordnung, DOT, technische Daten, Zustand, Kauf und Einlagerung; eigene Fahrzeuge und freigegebene Fahrzeuge verknüpfbar.
- Workspace: Fahrzeugliste mit Bearbeitung und Hinzufügen; Fahrerübersicht mit Suche und bedarfsgesteuertem Formular.
- Datenimport unter Backup & Umzug zusammengeführt; Reifen in ZIP-Backups; kompatible Backups ohne Reifentabelle bleiben importierbar (vorhandene Reifen werden dabei ersetzt).

## v1.1.6 – ZIP-Backup und Serverumzug (21. November 2026)

- Portabler Export und Restore von SQLite-Daten und zulässigen Uploads inklusive Manifest und Prüfsummen.
- Admin-only, Vorschau, Größen- und Schema-Prüfung, Bestätigung, lokale Rettungskopie, Transaktion und Recovery-Journal.
- Keine Passwort-Hashes oder Secrets im Export; Ziel-Admin behält sein Passwort, weitere Benutzer benötigen neue Passwörter.
- Benannte Volumes für Neuinstallationen; bestehende Mounts beibehalten.
- ZIP-Export und Import praktisch getestet; umfassendes Sicherheits-Audit noch offen.

## v1.1.5 – Fahrzeugmigration und UI-Verbesserungen (17. November 2026)

### Neu
- Migrationsassistent für Daten aus dem Original-Hammond: sichere SQLite-Datei `hammond.db` hochladen, Vorschau ansehen und Import gezielt bestätigen.
- Fahrzeuge, Tankungen und Ausgaben können übernommen werden; die Originaldatenbank wird schreibgeschützt verarbeitet.
- Gleiche Fahrzeugdatensätze aus der Originalinstallation werden anhand von Name, Kennzeichen, Hersteller und Modell gruppiert. Die zugehörige Tankhistorie bleibt erhalten.
- **Checkboxen** zur Auswahl einzelner Fahrzeuge vor dem Import.
- Duplikatprüfung anhand der ursprünglichen Eintrags-IDs, damit bereits importierte Tankungen und Ausgaben nicht mehrfach angelegt werden.
- Dokumentation für Installation, Update, Backup und Datenmigration.

### Verbesserungen
- Datenimport-Oberfläche für Dark Mode und Mobilgeräte überarbeitet; verständlichere Upload-Fehlermeldungen.
- SQLite-Upload im Browser auf natives Multipart-Formular umgestellt.
- Fahrzeugdetails: aufgeräumte Aktionsbuttons, mobile Tabellen und kontrastreiche Hover-Zustände.
- Vereinheitlichte Währungsbeträge und auf eine Nachkommastelle gerundete Literanzeigen.
- Wartungscenter und Dashboard-Erweiterungen aus dem bisherigen Entwicklungsstand konsolidiert.

### Noch offen
- **Tankbeleg-OCR** arbeitet nicht zuverlässig; manuelle Kontrolle notwendig.
- **Vollständiger Datenimport** ist noch nicht mit einer produktiven Original-Hammond-Datenbank auf Abschluss geprüft. Die Vorschau und Gruppierung wurden praktisch getestet.
- Anhänge, Rechnungsdateien, Nutzerkonten und Nutzerfreigaben werden nicht migriert.
- Kein vollständiges Sicherheits-Audit.

## v1.1.0 – Wartung, Kosten und mobile Darstellung
- Wartungsplanung für HU/TÜV, Inspektion und Service nach Datum und/oder Kilometerstand.
- Anzeige anstehender Wartungen im Dashboard.
- Erweiterte Kostenübersichten und Verlauf.
- „Tankung hinzufügen“ mit Fahrzeugauswahl; frühere Quick Entries aus der Oberfläche entfernt.
- Verbesserungen für Smartphone-Layouts, Dark Mode, Tabellen, Währungs- und Mengenformatierung.
- Erste experimentelle Tankbeleg-Texterkennung.

## v1.0.0 – Erste Stable-Version
- Modernisiertes Hammond-Layout mit responsiver Navigation und Dark-/Light-Mode.
- Fahrzeuge, Tankungen, Fahrer, Ausgaben, Werkstatt-Rechnungen und Tankstellensuche.
- Docker-Compose-Bereitstellung für einen lokalen Build und persistent gespeicherte Daten.

Weitere technische Änderungen stehen im [Git-Verlauf](https://github.com/Skoun181/hammond-modern/commits/develop/hammond-modern-next).
