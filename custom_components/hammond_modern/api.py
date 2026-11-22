"""Read-only Hammond Modern maintenance API client."""
from __future__ import annotations

import asyncio
from urllib.parse import urlsplit, urlunsplit

import aiohttp
from homeassistant.helpers.aiohttp_client import async_get_clientsession

from .const import CONF_ACCOUNT_ID

API_PATH = "/api/integrations/homeassistant/maintenance"


class InvalidAddress(ValueError):
    """Invalid Hammond instance address."""


class InvalidToken(Exception):
    """The API key was revoked or not accepted."""


class CannotConnect(Exception):
    """Cannot reach a compatible Hammond instance."""


def normalize_url(value: str) -> str:
    """Validate the operator-provided server URL, preserving proxy prefixes."""
    parts = urlsplit(value.strip())
    if (
        parts.scheme not in ("http", "https")
        or not parts.hostname
        or parts.username is not None
        or parts.password is not None
        or parts.query
        or parts.fragment
        or any(part in (".", "..") for part in parts.path.split("/"))
    ):
        raise InvalidAddress
    return urlunsplit((parts.scheme, parts.netloc, parts.path.rstrip("/"), "", ""))


async def async_fetch_maintenance(hass, address: str, token: str) -> dict:
    """Fetch the current tasks, without following redirects or logging secrets."""
    url = address + API_PATH
    try:
        session = async_get_clientsession(hass)
        async with session.get(
            url,
            headers={"Authorization": "Bearer " + token, "Accept": "application/json"},
            timeout=aiohttp.ClientTimeout(total=12),
            allow_redirects=False,
        ) as response:
            if response.status in (401, 403):
                raise InvalidToken
            if response.status != 200:
                raise CannotConnect
            payload = await response.json()
    except InvalidToken:
        raise
    except (aiohttp.ClientError, asyncio.TimeoutError, ValueError) as exc:
        raise CannotConnect from exc

    if not isinstance(payload, dict):
        raise CannotConnect
    if (
        payload.get("version") != 1
        or not isinstance(payload.get(CONF_ACCOUNT_ID), str)
        or not payload[CONF_ACCOUNT_ID]
        or not isinstance(payload.get("summary"), dict)
        or not isinstance(payload.get("tasks"), list)
    ):
        raise CannotConnect
    return payload
