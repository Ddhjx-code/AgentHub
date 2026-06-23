import zh from './zh'
import en from './en'
import type { Translations } from './zh'

export type Locale = 'zh' | 'en'
export type { Translations }

export const locales: Record<Locale, Translations> = { zh, en }
