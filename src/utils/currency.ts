import type { Currency } from '../types'

const CURRENCY_SYMBOLS: Record<Currency, string> = {
  USD: '$',
  GBP: '£',
  EUR: '€',
  NGN: '₦',
  KES: 'KSh',
}

export function getCurrencySymbol(currency: Currency): string {
  return CURRENCY_SYMBOLS[currency]
}

export function formatAmount(amount: string | number, currency: Currency): string {
  const num = typeof amount === 'string' ? parseFloat(amount) : amount
  const symbol = getCurrencySymbol(currency)
  return `${symbol}${num.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

export const SUPPORTED_CURRENCIES: Currency[] = ['USD', 'GBP', 'EUR', 'NGN', 'KES']
