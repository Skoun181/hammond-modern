# Home Assistant – Hammond Modern Wartungen (Develop 1.1.11-dev)

Diese Custom-Integration holt die **offenen Wartungen** deiner Fahrzeuge aus Hammond Modern in Home Assistant. **Nur lesender Zugriff**: keine Schreibzugriffe, keine Tankdaten, keine Kosten und keine Passwörter. Neue Wartungen werden bei der nächsten Abfrage automatisch als Sensor angelegt.

Die Integration liegt im selben GitHub-Repository unter `custom_components/hammond_modern` und wird als Custom Component installiert. Sie ist **keine offiziell im Home Assistant Marketplace gelistete Integration**.

## Voraussetzungen

- Hammond Modern **1.1.11-dev** oder eine neuere Version, in der der Home-Assistant-API-Endpunkt vorhanden ist.
- Home Assistant mit Zugang zu Hammond über das lokale Netzwerk oder einen abgesicherten Reverse Proxy.
- In Hammond ein Benutzer mit Zugriff auf mindestens ein Fahrzeug, für das Wartungen angelegt werden können.
- Für den ersten Test: die **separate Develop-Instanz** von Hammond verwenden. Stable und Develop nicht mit demselben Datenverzeichnis starten.

## 1. API-Schlüssel in Hammond erstellen

1. Hammond Modern öffnen und anmelden.
2. Links **Home Assistant** auswählen.
3. **Schlüssel erstellen** auswählen und den angezeigten Schlüssel direkt kopieren. Er beginnt mit `hmha_` und ist **nur einmal sichtbar**.
4. Falls du den Schlüssel verlierst: **Schlüssel erneuern**. Der alte verliert sofort seine Gültigkeit. **Schlüssel widerrufen** sperrt den Zugang vollständig.

Dieser Schlüssel ist **nicht** dein normaler Benutzer-Login, dein JWT_SECRET oder ein Home-Assistant-Token. Er erlaubt ausschließlich das Lesen der **eigenen** aktiven Wartungen. Wenn der Hammond-Benutzer deaktiviert wird oder ein Fahrzeug nicht mehr freigegeben ist, wird der entsprechende Zugriff ebenfalls blockiert.

**Wichtig:** Ein portables Hammond-ZIP enthält absichtlich **keine** Home-Assistant-Schlüssel. Bei einem Serverumzug oder Restore die Schlüssel in Hammond erneut ausstellen und in Home Assistant hinterlegen.

## 2. Integration nach Home Assistant kopieren

Der unterstützte Installationsweg für die Testversion ist **manuell**:

1. GitHub-Repository auf [Develop](https://github.com/Skoun181/hammond-modern/tree/develop/hammond-modern-next) öffnen.
2. **Den vollständigen Ordner `custom_components/hammond_modern`** kopieren.
3. In Home Assistants Konfigurationsordner einfügen, sodass die Struktur so aussieht:

```text
/config/
└── custom_components/
    └── hammond_modern/
        ├── __init__.py
        ├── manifest.json
        ├── api.py
        ├── config_flow.py
        ├── sensor.py
        ├── const.py
        ├── strings.json
        └── translations/
            ├── de.json
            └── en.json
```

4. Home Assistant neu starten.
5. **Einstellungen → Geräte & Dienste → Integration hinzufügen → Hammond Modern** öffnen.
6. Die **Hammond-Adresse** und den **API-Schlüssel** eingeben.

### Welche Adresse?

Die Adresse muss **aus dem Home-Assistant-Container** erreichbar sein. Bei einer Synology/macvlan-Installation kann das z. B. `http://192.168.0.36:3000` sein, **wenn** dein Hammond-Container dort tatsächlich auf Port 3000 erreichbar ist. Verwende sonst die passende LAN-URL oder die HTTPS-Adresse deines Reverse Proxy, ohne `/api` am Ende. Port, IP und DNS sind abhängig von deiner Installation.

Für Zugriffe über nicht vertrauenswürdige Netze unbedingt **HTTPS**, Zugriffsschutz und möglichst ein privates Netz/VPN nutzen. Niemals den Schlüssel in einer URL oder in öffentlich lesbaren Dateien ablegen.

## 3. Entitäten in Home Assistant

Die Integration legt ein Gerät `Hammond Modern` mit folgenden Sensoren an:

| Sensor | Wert |
|---|---|
| Wartungen offen | Gesamtzahl der offenen Wartungen |
| Wartungen überfällig | Datum oder Kilometerstand überschritten |
| Wartungen heute fällig | Datum oder km-Ziel genau erreicht |
| Wartungen bald fällig | Innerhalb der konfigurierten Vorwarnung |
| Nächste Wartung | Titel der dringendsten Wartung samt Attributen |
| Pro Wartung ein Statussensor | `Überfällig`, `Heute fällig`, `Bald fällig`, `Geplant` |

Jeder Wartungssensor enthält als Attribute Fahrzeug, Datum, Ziel-Kilometerstand, zuletzt erfassten km-Stand sowie verbleibende Tage oder Kilometer. Die **Vorwarnungseinstellungen aus Hammond** gelten unverändert.

Die Aktualisierung erfolgt **alle fünf Minuten**. Bei erledigten/gelöschten Wartungen wird der jeweilige Sensor **nicht verfügbar**, während neu hinzugefügte Wartungen automatisch erkannt werden. Ab Home Assistant können dann eigene Automationen und Dashboard-Karten gebaut werden. Die automatisch vergebenen Entity-IDs können abhängig von deiner HA-Installation abweichen.

**Beispielautomation** (Entity-ID nach dem Einrichten aus Home Assistant übernehmen):

```yaml
alias: Fahrzeugwartung überfällig
triggers:
  - trigger: numeric_state
    entity_id: sensor.DEINE_HAMMOND_WARTUNGEN_UEBERFAELLIG
    above: 0
actions:
  - action: persistent_notification.create
    data:
      title: "Fahrzeugwartung"
      message: "Mindestens eine Wartung ist überfällig. Schau in Hammond Modern."
mode: single
```

Du kannst auch die Statussensoren einzelner Wartungen in eine Entities-Card aufnehmen.

## 4. Verbindung testen

Auf einem **vertrauenswürdigen** System lässt sich der API-Endpunkt testen:

```bash
read -rs HA_KEY
curl -fsS -H "Authorization: Bearer $HA_KEY" \
  "http://DEINE-HAMMOND-ADRESSE:3000/api/integrations/homeassistant/maintenance"
unset HA_KEY
```

Erwartet wird `{"version":1,"account_id":"...","summary":{...},"tasks":[...]}`.

- **401:** Schlüssel fehlt, ist ungültig, widerrufen oder der Benutzer ist deaktiviert. In Hammond einen neuen erstellen, dann in Home Assistant die erneute Authentifizierung durchführen.
- **Verbindung fehlgeschlagen:** URL, Port und Erreichbarkeit vom Home-Assistant-Host prüfen; bei Synology auf macvlan-Netzwerk und Firewall achten.
- **0 Wartungen:** In Hammond unter **Wartungscenter** offene Wartungen anlegen und prüfen, ob der Schlüssel zu genau dem Benutzer gehört, der die Wartungen angelegt hat.
- **Fälligkeit nach km unbekannt:** Einen aktuellen km-Tank-/Ausgabeneintrag in Hammond erfassen. Meilen werden nicht mit km vermischt.

## API-Vertrag

`GET /api/integrations/homeassistant/maintenance`, Header `Authorization: Bearer hmha_...`

Ausgabe: `version`, `account_id`, `summary` mit `total, overdue, due, soon, planned` und `tasks` mit `id, vehicle_id, vehicle_name, title, severity, due_date, due_odometer, current_odometer, days_left, km_left`. Maximal 500 offene Wartungen pro Benutzer werden zurückgegeben. Der Endpunkt ist **schreibgeschützt** und benötigt kein JWT.

Der Token wird nur als SHA-256-Hash gespeichert, kann vom Nutzer rotiert oder widerrufen werden und wird niemals in portable ZIP-Backups übernommen. Nach dem Restore ist eine neue Verbindung erforderlich.

## Entwicklungsstand

- Diese Erweiterung wurde für **Develop** erstellt, nicht für den aktuellen Stable-Docker-Hub-Tag.
- Automatische CI: Go-Backend-Tests, Vue-Build, Python-Syntaxprüfung und grundlegende Manifest-/Übersetzungsvalidierung. Ein Test innerhalb einer vollständig installierten Home-Assistant-Instanz ist vor einem offiziellen Release sinnvoll.
