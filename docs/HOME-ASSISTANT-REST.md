# Home Assistant ohne Custom Integration (Develop)

Hammond Modern stellt bereits den **schreibgeschützten** Endpunkt
`GET /api/integrations/homeassistant/maintenance` bereit. Die Authentifizierung
erfolgt mit einem separaten `hmha_`-API-Schlüssel, **nicht** mit `JWT_SECRET`.

## Einmalige Einrichtung

1. In Hammond Modern (Develop) anmelden, links **Home Assistant** öffnen,
   **Schlüssel erstellen** und den nur einmal angezeigten Schlüssel kopieren.
2. Die Datei [hammond_modern.yaml](../homeassistant/hammond_modern.yaml) nach
   `/homeassistant/packages/hammond_modern.yaml` kopieren.
3. In `/homeassistant/configuration.yaml` **einmalig** sicherstellen, dass
   YAML-Packages geladen werden (vorhandenen `homeassistant:`-Block ergänzen,
   keinen zweiten Block anlegen):

   ```yaml
   homeassistant:
     packages: !include_dir_named packages
   ```

4. In `/homeassistant/secrets.yaml` nur den Schlüssel ergänzen:

   ```yaml
   hammond_modern_authorization: "Bearer hmha_DEIN_API_KEY"
   ```

   Das Präfix `Bearer ` ist erforderlich. Den API-Key niemals committen.
5. Alte Custom Integration `Hammond Modern` unter Geräte & Dienste entfernen,
   um doppelte Sensoren zu vermeiden. Anschließend YAML-Konfiguration prüfen
   und Home Assistant neu starten.

Die Paketdatei verwendet die **Develop-Testadresse**
`http://192.168.0.68:3007`. Bei anderer Adresse die `resource`-URL
in der Paketdatei ändern. Die Abfrage erfolgt alle 5 Minuten.

## Entitäten

- `sensor.hammond_modern_wartungen` – offene Wartungen gesamt,
  mit den Attributen `summary` und `tasks`
- `sensor.hammond_modern_uberfallig` – überfällig (Entity-ID kann durch
  Home Assistant abweichend vergeben werden)
- `sensor.hammond_modern_heute_fallig` – heute fällig
- `sensor.hammond_modern_bald_fallig` – bald fällig
- `sensor.hammond_modern_geplant` – geplant
- `sensor.hammond_modern_nachste_wartung` – nächste Wartung mit Attributen

**Hinweis:** Die Template-Sensoren verwenden intern
`sensor.hammond_modern_wartungen`. Falls diese Entity-ID bereits durch eine
alte Integration belegt ist und Home Assistant stattdessen eine `_2`-ID
vergibt, die alte Entität zuerst entfernen oder die Referenzen anpassen.

## Test und Fehlersuche

API direkt testen (nur im vertrauenswürdigen Netz):

```bash
read -rs HA_KEY
curl -fsS -H "Authorization: Bearer $HA_KEY" \
  http://192.168.0.68:3007/api/integrations/homeassistant/maintenance
unset HA_KEY
```

- HTTP 401: API-Key ungültig, widerrufen oder falscher Benutzer.
- Verbindungsfehler: Home Assistant muss die LAN-Adresse und Port 3007
  tatsächlich erreichen können (Firewall/Container-Netzwerk beachten).
- 0 Wartungen: prüfen, ob der angemeldete Hammond-Benutzer offene Wartungen besitzt.
- Zu große Attribute: Home Assistant begrenzt die Größe von Sensorzuständen;
  die REST-Antwort enthält maximal 500 Wartungen. Bei vielen Wartungen
  gegebenenfalls Attribute reduzieren.

Nur lesender Zugriff. Über ungesicherte Netze HTTPS/VPN nutzen.
