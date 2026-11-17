// Display-only formatting; calculations and stored values remain unchanged.
export function formatMoney(value, currency = 'EUR') {
  const number = Number(value)
  return new Intl.NumberFormat('de-DE', {
    style: 'currency', currency: currency || 'EUR'
  }).format(Number.isFinite(number) ? number : 0)
}

export function formatDecimal(value, digits = 2) {
  const number = Number(value)
  return new Intl.NumberFormat('de-DE', {
    minimumFractionDigits: digits, maximumFractionDigits: digits
  }).format(Number.isFinite(number) ? number : 0)
}

export function formatLiters(value) {
  return `${formatDecimal(value)} L`
}
