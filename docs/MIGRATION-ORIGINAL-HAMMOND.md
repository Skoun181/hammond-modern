# Original-Hammond nach Hammond Modern migrieren

## Umfang (erste Version)
Der Assistent unter **Datenimport → Original-Hammond übernehmen** importiert:
- Fahrzeuge (Hersteller, Modell, Kennzeichen, Kraftstoffart und Einheiten)
- Tankungen (Datum, Kilometerstand, Menge, Preis, Kosten, Volltankung, Tankstelle, Notizen)
- Ausgaben (Datum, Kosten, Kilometerstand, Art und Notizen; vorhandene Werkstatt-Metadaten, soweit im DB-Schema gespeichert)

**Nicht automatisch übertragen:** Benutzerkonten und Passwörter, andere Benutzerfreigaben, Fahrerzuordnungen, Dokumente und Rechnungsdateien aus dem Assets-Verzeichnis, Quick Entries, alte Erinnerungen. Die neue Wartungsplanung ist davon unabhängig.

## Vorher sichern
1. **Original-Hammond anhalten** (oder eine konsistente SQLite-Online-Sicherung erstellen). SQLite kann bei laufender Anwendung noch Daten in `hammond.db-wal` haben; die allein kopierte `hammond.db` ist dann möglicherweise unvollständig.
2. Die Datei `hammond.db` aus dem Original-Konfigurationsverzeichnis separat auf den lokalen Rechner kopieren.
3. **Hammond Modern ebenfalls sichern**: insbesondere `data/config/hammond.db` und `data/assets`. Sicherungen von aktiven SQLite-Datenbanken ebenfalls konsistent erstellen.
4. Original und Modern müssen **unterschiedliche** Konfigurations- und Asset-Verzeichnisse verwenden. Niemals eine alte Datenbankdatei direkt über die moderne kopieren.

## Ablauf
1. In Hammond Modern anmelden und im Menü **Datenimport** öffnen.
2. **Hammond-Migration starten** auswählen.
3. Originale `hammond.db` hochladen (bis 50 MB).
4. Vorschau ansehen: Anzahl Fahrzeuge, Tankungen und Ausgaben sowie eventuelle vorhandene Fahrzeuge. Identische Quellfahrzeuge mit gleichem Namen, Kennzeichen, Hersteller und Modell werden als ein Fahrzeug angezeigt; die Historie sämtlicher zugehöriger IDs wird zusammengeführt.
5. Per Checkbox gezielt auswählen, welche Fahrzeuge übernommen werden. Ohne Auswahl findet kein Import statt.
6. Nach eigener Prüfung den Import bestätigen.
7. Im Dashboard und in den Fahrzeugdetails stichprobenartig Daten prüfen.

## Verhalten und Sicherheit
- Die Quelldatenbank wird auf dem Server temporär und **schreibgeschützt** geöffnet; sie ersetzt nie die Modern-Datenbank.
- Ein Import läuft in einer Datenbanktransaktion. Bei einem technischen Fehler wird diese zurückgerollt.
- Daten werden dem aktuell angemeldeten Konto zugeordnet. Bestehende Fahrzeuge mit vollständig passender Identität (Nickname, Kennzeichen, Marke und Modell) werden wiederverwendet; andere werden neu angelegt.
- Wiederholte Tankungen und Ausgaben werden anhand ihrer ursprünglichen UUID oder einer Importquellenkennung übersprungen.
- Währungen und Distanzwerte der Quelle werden nicht automatisch umgerechnet.
- Anhänge aus der Original-Installation befinden sich nicht vollständig in der SQLite-Datei. Für eine Migration inklusive Dateien ist später eine gesonderte Funktion nötig.

## Grenzen
- Der Assistent erwartet die SQLite-Struktur von Hammond und ist **nicht** für beliebige SQLite-Dateien geeignet.
- Die Unterstützung für sehr alte oder angepasste Original-Hammond-Forks muss jeweils anhand einer Test-Sicherung geprüft werden.
- Diese Funktion ist Bestandteil von **v1.1.5**. Die Migrationsvorschau mit einem realen Original-Datenbestand wurde geprüft; ein vollständig abgeschlossener Import wurde noch nicht als erfolgreich getestet. Bitte zuerst mit Kopien und gesicherter Modern-Datenbank testen.
