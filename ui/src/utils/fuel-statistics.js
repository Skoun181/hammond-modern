// Tankmengen im aktuellen Monat/Jahr; Verbrauch nur Volltank-zu-Volltank.
// Alle Tankungen zwischen zwei gültigen Volltankungen fließen in den Zähler ein.
export function vehicleFuelStatistics(fillups, now = new Date()) {
  const items = (fillups || []).filter(f => f && Number.isFinite(Number(f.fuelQuantity)))
  const month = items.filter(f => { const d = new Date(f.date); return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth() })
  const year = items.filter(f => new Date(f.date).getFullYear() === now.getFullYear())
  const sum = data => data.reduce((total, f) => total + Number(f.fuelQuantity || 0), 0)
  const ordered = [...items].sort((a, b) => new Date(a.date) - new Date(b.date))
  let previousFull = null
  let litersSinceFull = 0
  let weightedLiters = 0
  let weightedDistance = 0
  for (const f of ordered) {
    const odo = Number(f.odoReading)
    if (!Number.isFinite(odo) || odo <= 0) { previousFull = null; litersSinceFull = 0; continue }
    if (f.hasMissedFillup === true) { previousFull = null; litersSinceFull = 0 }
    if (previousFull) litersSinceFull += Number(f.fuelQuantity || 0)
    if (f.isTankFull === true) {
      if (previousFull && odo > previousFull && litersSinceFull > 0) {
        weightedLiters += litersSinceFull
        weightedDistance += odo - previousFull
      }
      previousFull = odo
      litersSinceFull = 0
    }
  }
  return {
    monthLiters: sum(month), yearLiters: sum(year),
    consumption: weightedDistance > 0 ? (weightedLiters / weightedDistance) * 100 : null,
  }
}
