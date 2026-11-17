// Display-only formatting; calculations and stored values remain unchanged.
// Some legacy Hammond installations store display labels such as "euros (EUR)"
// instead of ISO 4217 currency codes. Intl.NumberFormat only accepts the code.
export function normalizeCurrency(currency) {
  if (typeof currency !== 'string') return 'EUR'
  const input = currency.trim()
  if (!input) return 'EUR'
  const code = input.match(/\(([A-Za-z]{3})\)\s*$/)
  const candidate = (code ? code[1] : input).toUpperCase()
  if (!/^[A-Z]{3}$/.test(candidate)) return 'EUR'
  try {
    Intl.NumberFormat('de-DE', { style: 'currency', currency: candidate })
    return candidate
  } catch (error) {
    return 'EUR'
  }
}

export function formatMoney(value, currency = 'EUR') {
  const number = Number(value)
  return new Intl.NumberFormat('de-DE', {
    style: 'currency',
    currency: normalizeCurrency(currency),
  }).format(Number.isFinite(number) ? number : 0)
}

export function formatDecimal(value, digits = 2) {
  const number = Number(value)
  return new Intl.NumberFormat('de-DE', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(Number.isFinite(number) ? number : 0)
}

export function formatLiters(value) {
  return `${formatDecimal(value)} L`
}
