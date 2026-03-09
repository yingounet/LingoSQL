import { shallowRef } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  setLocale as setI18nLocale,
  getInitialLocale,
  SUPPORTED_LOCALES,
  type AppLocale,
} from '@/i18n'
import type { Language } from 'element-plus/es/locale'

// Element Plus locale 映射
export const ELEMENT_LOCALE_MAP: Record<AppLocale, () => Promise<{ default: Language }>> = {
  en: () => import('element-plus/es/locale/lang/en.mjs'),
  'zh-Hans': () => import('element-plus/es/locale/lang/zh-cn.mjs'),
  'zh-Hant': () => import('element-plus/es/locale/lang/zh-tw.mjs'),
  ja: () => import('element-plus/es/locale/lang/ja.mjs'),
  fr: () => import('element-plus/es/locale/lang/fr.mjs'),
  de: () => import('element-plus/es/locale/lang/de.mjs'),
  es: () => import('element-plus/es/locale/lang/es.mjs'),
  ko: () => import('element-plus/es/locale/lang/ko.mjs'),
  ru: () => import('element-plus/es/locale/lang/ru.mjs'),
  it: () => import('element-plus/es/locale/lang/it.mjs'),
  pl: () => import('element-plus/es/locale/lang/pl.mjs'),
  nl: () => import('element-plus/es/locale/lang/nl.mjs'),
}

// 全局 Element Plus 语言包（供 App.vue 的 el-config-provider 使用）
export const elementLocale = shallowRef<Language | undefined>(undefined)

async function loadElementLocale(locale: AppLocale) {
  const loader = ELEMENT_LOCALE_MAP[locale]
  if (loader) {
    const mod = await loader()
    elementLocale.value = mod.default
  } else {
    const mod = await ELEMENT_LOCALE_MAP.en()
    elementLocale.value = mod.default
  }
}

export function useLocale() {
  const { locale, t } = useI18n()

  async function changeLocale(newLocale: AppLocale) {
    setI18nLocale(newLocale)
    await loadElementLocale(newLocale)
  }

  return {
    locale,
    t,
    changeLocale,
    supportedLocales: SUPPORTED_LOCALES,
    elementLocale,
  }
}

export { getInitialLocale, SUPPORTED_LOCALES, loadElementLocale }
export type { AppLocale }
