"""Summary and per-maintenance sensors for Hammond Modern."""
from __future__ import annotations

from homeassistant.config_entries import ConfigEntry
from homeassistant.core import HomeAssistant, callback
from homeassistant.helpers.entity import DeviceInfo
from homeassistant.helpers.entity_platform import AddEntitiesCallback
from homeassistant.helpers.update_coordinator import CoordinatorEntity
from homeassistant.components.sensor import SensorEntity

from . import HammondCoordinator
from .const import DOMAIN

SUMMARY_SENSORS = {
    "total": ("Wartungen offen", "mdi:car-wrench"),
    "overdue": ("Wartungen überfällig", "mdi:alert-octagon"),
    "due": ("Wartungen heute fällig", "mdi:calendar-alert"),
    "soon": ("Wartungen bald fällig", "mdi:calendar-clock"),
}
SEVERITY_LABELS = {
    "overdue": "Überfällig",
    "due": "Heute fällig",
    "soon": "Bald fällig",
    "planned": "Geplant",
}


def next_task(data: dict) -> dict | None:
    """API tasks arrive ordered by urgency and date."""
    tasks = data.get("tasks", [])
    return tasks[0] if tasks else None


class HammondSensor(CoordinatorEntity[HammondCoordinator], SensorEntity):
    """Shared integration device and update lifecycle."""

    _attr_has_entity_name = True

    def __init__(self, coordinator: HammondCoordinator, entry: ConfigEntry, suffix: str):
        super().__init__(coordinator)
        self._attr_unique_id = f"{entry.entry_id}_{suffix}"
        self._attr_device_info = DeviceInfo(
            identifiers={(DOMAIN, entry.entry_id)},
            name="Hammond Modern",
            manufacturer="Hammond Modern",
            model="Wartungsverwaltung",
            configuration_url=entry.data.get("url"),
        )


class HammondCountSensor(HammondSensor):
    """Number of maintenance tasks for a status category."""

    _attr_native_unit_of_measurement = None
    _attr_state_class = None

    def __init__(self, coordinator, entry, key: str):
        super().__init__(coordinator, entry, "count_" + key)
        self.key = key
        self._attr_name, self._attr_icon = SUMMARY_SENSORS[key]

    @property
    def native_value(self) -> int:
        return int(self.coordinator.data["summary"].get(self.key, 0))


class HammondNextSensor(HammondSensor):
    """Most urgent upcoming/current maintenance."""

    def __init__(self, coordinator, entry):
        super().__init__(coordinator, entry, "next")
        self._attr_name = "Nächste Wartung"
        self._attr_icon = "mdi:calendar-clock"

    @property
    def native_value(self) -> str:
        task = next_task(self.coordinator.data)
        if task is None:
            return "Keine"
        return str(task["title"])[:250]

    @property
    def extra_state_attributes(self) -> dict:
        task = next_task(self.coordinator.data)
        if task is None:
            return {}
        return {
            "fahrzeug": task.get("vehicle_name"),
            "status": task.get("severity"),
            "faellig_am": task.get("due_date"),
            "faellig_bei_km": task.get("due_odometer"),
            "aktuell_km": task.get("current_odometer"),
            "tage_verbleibend": task.get("days_left"),
            "km_verbleibend": task.get("km_left"),
        }


class HammondTaskSensor(HammondSensor):
    """A stable sensor for one maintenance task, discovered automatically."""

    def __init__(self, coordinator, entry, task: dict):
        self.task_id = task["id"]
        super().__init__(coordinator, entry, "task_" + self.task_id)
        self._attr_name = f"{task.get('vehicle_name', 'Fahrzeug')} – {task.get('title', 'Wartung')}"
        self._attr_icon = "mdi:car-wrench"

    def task(self) -> dict | None:
        return next(
            (item for item in self.coordinator.data.get("tasks", [])
             if item.get("id") == self.task_id),
            None,
        )

    @property
    def available(self) -> bool:
        return super().available and self.task() is not None

    @property
    def native_value(self) -> str | None:
        task = self.task()
        return SEVERITY_LABELS.get(task.get("severity"), "Geplant") if task else None

    @property
    def extra_state_attributes(self) -> dict:
        task = self.task()
        if task is None:
            return {}
        return {
            "fahrzeug": task.get("vehicle_name"),
            "status": task.get("severity"),
            "faellig_am": task.get("due_date"),
            "faellig_bei_km": task.get("due_odometer"),
            "aktuell_km": task.get("current_odometer"),
            "tage_verbleibend": task.get("days_left"),
            "km_verbleibend": task.get("km_left"),
        }


async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    coordinator = hass.data[DOMAIN][entry.entry_id]
    entities: list[SensorEntity] = [
        HammondCountSensor(coordinator, entry, key) for key in SUMMARY_SENSORS
    ]
    entities.append(HammondNextSensor(coordinator, entry))
    async_add_entities(entities)

    known: set[str] = set()

    @callback
    def discover_tasks() -> None:
        new_entities: list[HammondTaskSensor] = []
        for task in coordinator.data.get("tasks", []):
            task_id = task.get("id")
            if isinstance(task_id, str) and task_id and task_id not in known:
                known.add(task_id)
                new_entities.append(HammondTaskSensor(coordinator, entry, task))
        if new_entities:
            async_add_entities(new_entities)

    discover_tasks()
    entry.async_on_unload(coordinator.async_add_listener(discover_tasks))
