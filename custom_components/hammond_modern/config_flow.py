"""User interface configuration and key rotation for Hammond Modern."""
from __future__ import annotations

import voluptuous as vol

from homeassistant import config_entries
from homeassistant.config_entries import ConfigFlowResult
from homeassistant.const import CONF_URL

from .api import CannotConnect, InvalidAddress, InvalidToken, async_fetch_maintenance, normalize_url
from .const import CONF_ACCOUNT_ID, CONF_API_KEY, DOMAIN


class HammondModernConfigFlow(config_entries.ConfigFlow, domain=DOMAIN):
    """Set up the instance using its separate read-only key."""

    VERSION = 1

    async def async_step_user(self, user_input=None) -> ConfigFlowResult:
        errors = {}
        if user_input is not None:
            try:
                address = normalize_url(user_input[CONF_URL])
                payload = await async_fetch_maintenance(
                    self.hass, address, user_input[CONF_API_KEY]
                )
            except InvalidAddress:
                errors["base"] = "invalid_url"
            except InvalidToken:
                errors["base"] = "invalid_auth"
            except CannotConnect:
                errors["base"] = "cannot_connect"
            else:
                await self.async_set_unique_id(
                    address.lower() + "|" + payload[CONF_ACCOUNT_ID]
                )
                self._abort_if_unique_id_configured()
                return self.async_create_entry(
                    title="Hammond Modern",
                    data={
                        CONF_URL: address,
                        CONF_API_KEY: user_input[CONF_API_KEY],
                        CONF_ACCOUNT_ID: payload[CONF_ACCOUNT_ID],
                    },
                )

        return self.async_show_form(
            step_id="user",
            data_schema=vol.Schema(
                {
                    vol.Required(CONF_URL): str,
                    vol.Required(CONF_API_KEY): str,
                }
            ),
            errors=errors,
        )

    async def async_step_reauth(self, entry_data: dict) -> ConfigFlowResult:
        """Reauthenticate after the user rotates/revokes the Hammond key."""
        return await self.async_step_reauth_confirm()

    async def async_step_reauth_confirm(self, user_input=None) -> ConfigFlowResult:
        errors = {}
        entry = self._get_reauth_entry()
        if user_input is not None:
            try:
                payload = await async_fetch_maintenance(
                    self.hass, entry.data[CONF_URL], user_input[CONF_API_KEY]
                )
            except InvalidToken:
                errors["base"] = "invalid_auth"
            except CannotConnect:
                errors["base"] = "cannot_connect"
            else:
                if payload[CONF_ACCOUNT_ID] != entry.data[CONF_ACCOUNT_ID]:
                    errors["base"] = "different_account"
                else:
                    return self.async_update_reload_and_abort(
                        entry, data_updates={CONF_API_KEY: user_input[CONF_API_KEY]}
                    )
        return self.async_show_form(
            step_id="reauth_confirm",
            data_schema=vol.Schema({vol.Required(CONF_API_KEY): str}),
            errors=errors,
        )
