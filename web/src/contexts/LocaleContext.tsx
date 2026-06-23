import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'
import type { Locale, Translations } from '../i18n'
import { locales } from '../i18n'

interface LocaleState {
  locale: Locale
  t: Translations
  setLocale: (l: Locale) => void
  toggleLocale: () => void
}

const LocaleContext = createContext<LocaleState | null>(null)

export function LocaleProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(() => {
    const stored = localStorage.getItem('locale')
    return (stored === 'en' || stored === 'zh') ? stored : 'zh'
  })

  const setLocale = useCallback((l: Locale) => {
    setLocaleState(l)
    localStorage.setItem('locale', l)
  }, [])

  const toggleLocale = useCallback(() => {
    setLocaleState((prev) => {
      const next = prev === 'zh' ? 'en' : 'zh'
      localStorage.setItem('locale', next)
      return next
    })
  }, [])

  return (
    <LocaleContext.Provider value={{ locale, t: locales[locale], setLocale, toggleLocale }}>
      {children}
    </LocaleContext.Provider>
  )
}

export function useLocale() {
  const ctx = useContext(LocaleContext)
  if (!ctx) throw new Error('useLocale must be used within LocaleProvider')
  return ctx
}
