# Fahrzeug- und Tankstellensuche testen

Stand: Entwicklungsbranch `develop/hammond-modern-next`.

## Änderungen

- Fahrzeugmodelle werden über `GET /api/vehicle-models?make=BMW` vom Go-Backend abgefragt. Browser-CORS entfällt. Ergebnisse werden im Prozess 24 Stunden zwischengespeichert; manuelle Eingabe bleibt möglich.
- Tankstellen: `GET /api/stations?plz=69115` nutzt Nominatim für PLZ-Koordinaten und Tankerkönig für Stationen (Radius 5 km). Ergebnisse bleiben 15 Minuten im Cache. Zwischen externen PLZ-Anfragen gelten mindestens zwei Sekunden Abstand.
- Für die Tankstellensuche ist **ein eigener** `TANKERKOENIG_API_KEY` erforderlich. Den Key nur in `.env` speichern, nicht committen.

## Debian-LXC-Installation

Im vorhandenen Git-Checkout:

```bash
git switch develop/hammond-modern-next
git pull --ff-only origin develop/hammond-modern-next
```

In `.env` müssen `JWT_SECRET` und optional `TANKERKOENIG_API_KEY` gesetzt sein. Ein Secret bitte nicht bei jedem Rebuild neu erzeugen.

```bash
docker compose -f compose.test.yml up -d --build
docker compose -f compose.test.yml ps
docker compose -f compose.test.yml logs --tail=100
```

Browser: `http://LXC-IP:3006`.

## Funktionstest

1. Anmelden und unter **Fahrzeug hinzufügen** z. B. BMW eingeben; Modelle sollten als Vorschläge erscheinen.
2. Wenn keine Modelle verfügbar sind, eigene Modellbezeichnung eingeben. Das Speichern muss weiterhin möglich sein.
3. Bei **Tankung hinzufügen → Weitere Angaben** PLZ `69115` eingeben und **Suchen** klicken.
4. Ohne API-Key muss eine erklärende Fehlermeldung erscheinen. Mit gültigem Key sollten Stationen im Umkreis von ungefähr fünf Kilometern angeboten werden, sofern der Anbieter Ergebnisse liefert.
5. Bei externen Störungen Logs und Browser-Netzwerkanfragen prüfen. HTTP 401/403 beim Anbieter oder leere Ergebnislisten sind nicht automatisch Frontend-Fehler.

**Achtung:** Die NHTSA-Modellliste ist US-orientiert; europäische Modellvarianten können fehlen. Das ist eine Datenabdeckungsgrenze, kein Ladefehler. Der Backend-Build und die End-to-End-Funktionen müssen auf der Testinstallation bestätigt werden.
