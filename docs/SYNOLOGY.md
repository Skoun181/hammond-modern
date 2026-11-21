# Hammond Modern v1.1.6 – Synology DSM / Dockhand

## Status

Das Docker-Hub-Image `skoun181/hammond-modern:1.1.5` wurde per GitHub Actions erfolgreich gebaut und auf einer Synology mit Dockhand als Compose-Stack gestartet (Anwenderbestätigung). Diese Bestätigung bezieht sich auf v1.1.5. Für v1.1.6 wird ein erneuter DSM-Test empfohlen; ein umfassender Sicherheits-Audit steht aus.

**Voraussetzung:** Synology mit AMD64/x86-64, Docker/Container Manager und Dockhand. Das aktuelle Image ist nicht für ARM gebaut.

## Installation per Docker Hub (empfohlen)

1. In Dockhand einen **normalen Compose-Stack** anlegen (kein Git-Stack und kein lokaler Build erforderlich).
2. Die Datei [`docker-compose.dockerhub.yml`](../docker-compose.dockerhub.yml) als Vorlage verwenden. Sie enthält das veröffentlichte Image und die bereits getestete Synology-Konfiguration:
   - Image: `skoun181/hammond-modern:1.1.6`
   - Externes Docker-Netzwerk `macvlan`
   - Container-IP: `192.168.0.36`, Anwendung auf Port `3000`
   - Persistente Daten: `/volume2/docker/Hammond/Hammond-config` und `/volume2/docker/Hammond/Hammond-assets`
3. **Netzwerkname, Subnetz, IP und Verzeichnisse** an die eigene Synology anpassen. `macvlan` muss bereits existieren; die gewählte IP muss frei sein.
4. In Dockhand als Environment-Variablen `JWT_SECRET` (langes zufälliges Secret) und optional `TANKERKOENIG_API_KEY` setzen. Secrets nicht ins öffentliche GitHub-Repository übernehmen.
5. Stack deployen und von einem anderen Gerät im LAN `http://192.168.0.36:3000` öffnen.

Mit macvlan hat der Container eine eigene IP; `ports: 3000:3000` ist dafür nicht erforderlich. Die Synology selbst kann aufgrund der macvlan-Isolation normalerweise nicht direkt über diese IP auf den Container zugreifen.

## Alternative ohne macvlan (Port 3006)

Wenn kein macvlan gewünscht ist, können die `networks`-Blöcke entfernt und stattdessen am Service diese Ports eingetragen werden:

```yaml
ports:
  - "3006:3000"
```

Dann erfolgt der Zugriff über `http://SYNOLOGY-IP:3006`.

## Aktualisierung und Datensicherung

- Vor jedem Update `/config` und `/assets` sichern; SQLite-Datenbank nur konsistent sichern (z. B. Container stoppen oder SQLite-Online-Backup).
- Bei einer neuen Version den Image-Tag in Dockhand ändern und den Stack neu deployen.
- Den bestehenden `JWT_SECRET` beibehalten.
- Keine Volumes oder Datenverzeichnisse löschen.
- Den HTTP-Port nicht ungeschützt ins Internet stellen; für externen Zugriff HTTPS/Reverse Proxy oder VPN verwenden.

## Selbst aus dem Quellcode bauen

Für Entwickler gibt es weiterhin `docker-compose.synology.yml` und den Git-Stack `docker-compose.dockhand.yml`. Für eine normale Synology-Installation ist der Pull des Docker-Hub-Images einfacher.

## Bekannte Grenzen

- Die Foto-OCR für Tankbelege bleibt experimentell.
- Die Original-Hammond-Migration wurde noch nicht vollständig auf produktiven Daten getestet.
- Das Projekt nutzt ältere Abhängigkeiten aus der Hammond-Codebasis; ein vollständiges Sicherheits-Audit fehlt.
