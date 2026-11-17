# Hammond Modern – Änderungsverlauf

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
