# Hammond Modern Next – Wartung, Statistik und Belegscan

**Stand:** Entwicklungsbranch `develop/hammond-modern-next`. **Nicht auf Stable übernommen und noch nicht auf dem Debian-LXC durchgebaut/getestet.**

## Neue Funktionen

### 1. Wartungscenter (ohne Benachrichtigungen)

- Navigation → **Wartungscenter**.
- Fahrzeug und Wartungstitel (z. B. HU/TÜV, Ölwechsel, Inspektion, Reifenwechsel) auswählen.
- Fälligkeitsdatum, Ziel-Kilometerstand oder beides eintragen.
- Notizen speichern; Einträge bearbeiten, als erledigt markieren oder löschen.
- Offene Einträge zeigen die verbleibenden Tage und die Distanz zur Ziel-Laufleistung, sofern mindestens ein Kilometerstand aus Tankungen bekannt ist.
- Kein Push, keine E-Mail, keine Automatisierung und kein automatisches Erzeugen von Folgeterminen.

### 2. Kosten und Statistik

- Dashboard → **Kostenentwicklung – letzte 6 Monate**.
- Farbige Monatsbalken: Orange für Tankungen, Violett für Ausgaben.
- Erfasste Gesamtkosten pro Fahrzeug und grober €/km-Wert, sofern zwei unterschiedliche Kilometerstände vorliegen.
- Der €/km-Wert ist eine vereinfachte Orientierung: alle gespeicherten Kosten dividiert durch die Spanne der gespeicherten Kilometerstände; nicht mit realen Gesamtkosten pro gefahrenem Kilometer verwechseln.
- Die bestehende Verbrauchsberechnung bleibt unverändert.

### 8. Belegfoto per Handy

- Fahrzeug → **Tankung hinzufügen** → **Beleg fotografieren / hochladen**.
- Smartphones können über den HTML-Dateidialog (`capture="environment"`) die Rückkamera öffnen, sofern der Browser das unterstützt. Alternativ JPG/PNG aus der Galerie auswählen.
- Bild bis 5 MB wird nur an den eigenen, angemeldeten Hammond-Server gesendet. Tesseract OCR (`deu+eng`) liest das Bild lokal.
- Vorschläge für Datum, Tankstelle, Liter, Literpreis und Gesamtbetrag werden nur in das Formular eingetragen, **nicht automatisch gespeichert**.
- Nach Kontrolle aller Angaben wird die Tankung über den bisherigen Speichern-Button angelegt.
- Belegbild wird nach der Texterkennung gelöscht; kein externer OCR-Dienst. Der Browser lädt das Foto nicht als dauerhaften Beleg hoch.
- Der OCR-Parser arbeitet mit Heuristiken; abhängig von Schriftart, Beleuchtung, Kassenbonlayout und OCR-Qualität können Werte fehlen oder falsch sein.
- Die klassische Tankstellensuche per PLZ funktioniert weiterhin unabhängig davon.

### Webseite auf dem Home-Bildschirm

Es wird bewusst **keine eigene Smartphone-App** ausgeliefert.

- **Android (Chrome):** Webseite öffnen → Drei-Punkte-Menü → `Zum Startbildschirm hinzufügen` (Bezeichnung je nach Version).
- **iPhone (Safari):** Webseite öffnen → Teilen → `Zum Home-Bildschirm`.
- Die Verknüpfung öffnet weiterhin die Webseite; keine Offline-Synchronisierung oder installierbare PWA wird versprochen.
- Kameradialoge hängen von Gerät und Browser ab. Für entfernten Zugriff nur HTTPS/VPN verwenden.

## Installieren und testen

Bestehenden Entwicklertest-Checkout verwenden:

```bash
cd /opt/docker/hammond-fork/hammond-modern-git
git switch develop/hammond-modern-next
git pull --ff-only origin develop/hammond-modern-next
docker compose -f compose.test.yml up -d --build
docker compose -f compose.test.yml ps
```

Hinweis: Beim Start wird `maintenance_tasks` über GORM AutoMigrate ergänzt. Eine Datenbankmigration wird nicht automatisch zurückgerollt, wenn der Code später auf Stable zurückgestellt wird. **Die bestehende .env und `data/` nicht löschen.**

### Testablauf

1. App öffnen; Navigation enthält **Wartungscenter**.
2. Neue Wartung mit Datum anlegen, nach Reload wiederfinden; bearbeiten, erledigen und löschen.
3. Weitere Wartung mit Ziel-Kilometerstand anlegen und Restkilometeranzeige prüfen.
4. Dashboard öffnen; Monatsbalken und Kosten je Fahrzeug prüfen.
5. Fahrzeug → Tankung hinzufügen → Kamera bzw. Bildauswahl öffnen und einen **Test-Kassenbon** als JPG/PNG verwenden.
6. Von OCR vorgeschlagene Werte ausdrücklich mit dem Bon vergleichen und gegebenenfalls korrigieren. Bei schlechter Erkennung manuell ausfüllen.
7. Tankung speichern und erneut das Dashboard laden. Prüfen, dass die Kosten und Liter aktualisiert wurden.

### Build/Fehleranalyse

```bash
docker compose -f compose.test.yml logs --tail=100
docker compose -f compose.test.yml build --progress=plain
```

Im Server-Build lassen sich zudem gezielte OCR-Parser-Tests ausführen:

```bash
cd server
go test ./controllers -run TestParseReceipt -v
```

**Hinweis:** In dieser Umgebung wurden keine vollständigen Docker-/Browser-End-to-End-Tests ausgeführt. Erst wenn die Tests auf dem Debian-LXC bestanden sind, sollte über eine Aktualisierung des Stable-Branches entschieden werden.

## Dashboard-/Mobil-Update und OCR-Nachbesserung

- Die Quick-Entry-Navigation, die Quick-Entry-Vue-Seiten und die Quick-Entry-API-Routen sind im Entwicklungsbranch entfernt. Die historischen SQLite-Tabellen bleiben zur Sicherheit bestehen; bestehende Anhänge werden nicht gelöscht.
- Im Dashboard steht stattdessen die Wartungsübersicht mit den nächsten vier offenen Terminen.
- Das Wartungscenter zeigt dieselbe Übersicht. Das Bearbeitungsformular erscheint erst nach „Neue Wartung hinzufügen“ oder „Bearbeiten“.
- Der Dashboard-Button für Tankungen bietet bei mehreren Fahrzeugen eine Fahrzeugauswahl.
- Schmale iPhone-Ansichten: kleinere Innenabstände, umbruchfähige Eingabezeilen, Karten und Buttons.
- Für iPhone-Bonfotos werden vor lokaler Tesseract-OCR zwei kontrastangepasste Versionen erstellt. Der Beleg wird weder in der App gespeichert noch an externe OCR-Server geschickt.
- OCR-Ergebnisse werden nur als Entwurf in das Formular eingetragen. Gerade das Datum unbedingt mit dem Beleg vergleichen: Eine optische Verwechslung von „15“ und „13“ ist möglich.
- Wenn andere Artikel auf dem Kassenbon stehen, berechnet die Erkennung den Kraftstoffbetrag aus Litern und Literpreis, anstatt die gesamte Einkaufssumme zu übernehmen.

### Beispiel: HEM-Beleg mit weiteren Einkäufen

Das für die Fehlersuche verwendete Beispiel zeigt 41,00 Liter zu 1,869 EUR/Liter (= 76,63 EUR Kraftstoffkosten), während die gesamte Kartenzahlung 78,24 EUR beträgt. Der Gesamtbetrag **darf nicht** als Tankkosten gespeichert werden.

Der tatsächliche Foto-Scan und die mobilen Layoutänderungen müssen auf der iPhone-/LXC-Testinstallation erst bestätigt werden; ein synthetischer Textparser-Test ist kein Ende-zu-Ende-Kameratest. Private Belegfotos bitte nicht ins öffentliche Repository committen.

