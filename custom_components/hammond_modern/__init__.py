"""Home Assistant integration for Hammond Modern maintenance."""
from __future__ import annotations

from datetime import timedelta
import logging

from homeassistant.config_entries import ConfigEntry
from homeassistant.const import CONF_URL, Platform
from homeassistant.core import HomeAssistant
from homeassistant.exceptions import ConfigEntryAuthFailed
from homeassistant.helpers.update_coordinator import DataUpdateCoordinator, UpdateFailed

from .api import CannotConnect, InvalidToken, async_fetch_maintenance
from .const import CONF_API_KEY, DOMAIN, UPDATE_INTERVAL_MINUTES

PLATFORMS = [Platform.SENSOR]
_LOGGER = logging.getLogger(__name__)


class HammondCoordinator(DataUpdateCoordinator[dict]):
    """Poll all maintenance entities in one request per instance."""

    def __init__(self, hass: HomeAssistant, entry: ConfigEntry) -> None:
        super().__init__(
            hass,
            _LOGGER,
            name="Hammond Modern Wartungen",
            update_interval=timedelta(minutes=UPDATE_INTERVAL_MINUTES),
        )
        self.entry = entry

    async def _async_update_data(self) -> dict:
        try:
            return await async_fetch_maintenance(
                self.hass, self.entry.data[CONF_URL], self.entry.data[CONF_API_KEY]
            )
        except InvalidToken as exc:
            raise ConfigEntryAuthFailed("Hammond Modern API key invalid or revoked") from exc
        except CannotConnect as exc:
            raise UpdateFailed("Hammond Modern maintenance API unavailable") from exc


async def async_setup_entry(hass: HomeAssistant, entry: ConfigEntry) -> bool:
    coordinator = HammondCoordinator(hass, entry)
    await coordinator.async_config_entry_first_refresh()
    hass.data.setdefault(DOMAIN, {})[entry.entry_id] = coordinator
    await hass.config_entries.async_forward_entry_setups(entry, PLATFORMS)
    return True


async def async_unload_entry(hass: HomeAssistant, entry: ConfigEntry) -> bool:
    success = await hass.config_entries.async_unload_platforms(entry, PLATFORMS)
    if success:
        hass.data[DOMAIN].pop(entry.entry_id, None)
    return success
