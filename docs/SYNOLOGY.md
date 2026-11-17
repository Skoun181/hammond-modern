# Hammond Modern – Synology DSM / Container Manager

## Status
The application source has been redesigned. The complete image build has **not** been verified in this environment. The original Vue 2 / Vue CLI 4 and Go dependencies remain legacy and may require additional fixes on current toolchains. Test on a non-public network before production use.

## Install
1. Create `/volume1/docker/hammond-modern/config` and `/volume1/docker/hammond-modern/assets` in DSM File Station.
2. Copy `.env.example` to `.env` and replace the placeholder with a long random secret (`openssl rand -hex 32`). Never commit `.env`.
3. Check that port 3005 is unused. Check that `/volume1` exists on your NAS; adjust the paths if needed.
4. In a terminal with Docker Compose available, run `docker compose -f docker-compose.synology.yml build` and then `docker compose -f docker-compose.synology.yml up -d`.
5. Open `http://NAS-IP:3005` on your local network and complete first-time setup.
6. In DSM Container Manager, you can alternatively create a Project using the Compose file, provided the Docker image has been built or can be built on that NAS.

## Security and data
- Do not publish port 3005 directly to the internet. Use a reverse proxy with HTTPS and access controls if remote access is needed.
- Back up `/config` (SQLite DB) and `/assets` regularly.
- The example PUID/PGID values are placeholders. Adjust to your Synology user/group IDs and folder permissions.
- Keep the AGPL license and copyright notices when redistributing modifications.

## Known modernization work
- Vue 2 and Vue CLI 4 remain unsupported upstream; migrating to Vue 3 and a supported component library is a separate high-risk task.
- The backend still uses the old JWT library and dependencies. Review security before exposing the service.
- The Dockerfile attempts newer Go/Node build stages, but a passing Docker build and end-to-end tests are still required.
