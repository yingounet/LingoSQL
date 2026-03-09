import { createI18n } from 'vue-i18n'
import type { Locale } from 'vue-i18n'
import en from './locales/en.json'
import zhHans from './locales/zh-Hans.json'
import zhHant from './locales/zh-Hant.json'
import ja from './locales/ja.json'
import fr from './locales/fr.json'
import de from './locales/de.json'
import es from './locales/es.json'
import ko from './locales/ko.json'
import ru from './locales/ru.json'
import it from './locales/it.json'
import pl from './locales/pl.json'
import nl from './locales/nl.json'

export type AppLocale = 
  | 'en' 
  | 'zh-Hans' 
  | 'zh-Hant' 
  | 'ja' 
  | 'fr' 
  | 'de' 
  | 'es' 
  | 'ko' 
  | 'ru' 
  | 'it' 
  | 'pl' 
  | 'nl'

export const SUPPORTED_LOCALES: { value: AppLocale; label: string }[] = [
  { value: 'en', label: 'English' },
  { value: 'zh-Hans', label: '简体中文' },
  { value: 'zh-Hant', label: '繁體中文' },
  { value: 'ja', label: '日本語' },
  { value: 'fr', label: 'Français' },
  { value: 'de', label: 'Deutsch' },
  { value: 'es', label: 'Español' },
  { value: 'ko', label: '한국어' },
  { value: 'ru', label: 'Русский' },
  { value: 'it', label: 'Italiano' },
  { value: 'pl', label: 'Polski' },
  { value: 'nl', label: 'Nederlands' },
]

const messages = {
  en,
  'zh-Hans': zhHans,
  'zh-Hant': zhHant,
  ja,
  fr,
  de,
  es,
  ko,
  ru,
  it,
  pl,
  nl,
}

const STORAGE_KEY = 'lingosql-locale'

function getStoredLocale(): AppLocale | null {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored && stored in messages) return stored as AppLocale
  } catch {}
  return null
}

function getBrowserLocale(): AppLocale | null {
  const browser = navigator.language
  if (!browser) return null
  const lang = browser.toLowerCase()
  if (lang.startsWith('zh')) {
    if (lang.includes('tw') || lang.includes('hk') || lang.includes('hant')) return 'zh-Hant'
    return 'zh-Hans'
  }
  const map: Record<string, AppLocale> = {
    'ja': 'ja', 'fr': 'fr', 'de': 'de', 'es': 'es', 'ko': 'ko',
    'ru': 'ru', 'it': 'it', 'pl': 'pl', 'nl': 'nl', 'en': 'en',
  }
  const base = lang.split('-')[0]
  return map[base] ?? null
}

export function getInitialLocale(): AppLocale {
  return getStoredLocale() ?? getBrowserLocale() ?? 'en'
}

export const i18n = createI18n({
  legacy: false,
  locale: getInitialLocale(),
  fallbackLocale: 'en',
  messages,
  globalInjection: true,
})

export function setLocale(locale: AppLocale) {
  i18n.global.locale.value = locale as Locale
  localStorage.setItem(STORAGE_KEY, locale)
}
