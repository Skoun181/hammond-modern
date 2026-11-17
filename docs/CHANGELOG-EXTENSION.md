# Hammond Modern – Erweiterung (ungetesteter Quellcode)

## Umgesetzt im Code

- Fahrer: eigene Tabelle, authentifizierte CRUD-API und Verwaltungsseite, Führerschein-Ablaufdatum als optionales Feld
- Tankungen: optionaler Fahrer, kilometerbasierte Verbrauchsstatistik, deutscher UI-Text
- Werkstattausgaben: Werkstatt, Arbeiten, PDF/JPG/PNG-Rechnung, Speicherung im geschützten Asset-Verzeichnis, verknüpft über Attachment-ID
- Autos: 110 suchbare Hersteller; Onlinedaten für Modelle über NHTSA vPIC mit manueller Alternative
- Tankstellen: PLZ-Abfrage über Nominatim + Tankerkönig (API-Key nötig, gedrosselt und gecacht)
- Dashboard: getankte Liter pro Fahrzeug und Monat/Jahr; Volltank-Verbrauch; globale Kostenübersicht für die dem Nutzer zugeordneten Fahrzeuge
- Oberfläche: mehrere deutsche Beschriftungen; **keine vollständige Lokalisierung aller Seiten**

## Offene Punkte für Freigabe

- Docker-Build und UI durch den Nutzer auf Debian-13-LXC durchführen (hier kein vollständiger Build mit npm/Go-Abhängigkeiten möglich)
- Datenmigration mit Kopie der alten DB testen; Datenbankmigrationen sind nicht reversibel
- Fahrerrechte über mehrere Hammond-Benutzer/Shared Vehicles spezifizieren und testen
- Mehrbenutzer-Zugriffe auf Fahrzeuginhalte im historischen Code auditieren
- Station API-Key konfigurieren, Nutzungsbedingungen beachten; nicht in Git committen
- Browser-Funktionstests inklusive Hochladen/Herunterladen, Fahrerbearbeitung und Fahrzeugmodellabfrage
- Langfristig alte Vue-2-/Backend-Abhängigkeiten ersetzen
