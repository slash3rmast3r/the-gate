import { createContext, ReactNode, useContext, useMemo, useState } from "react";
import { Lang, translations } from "./translations";

type I18nContextValue = {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: (key: string) => string;
};

const I18nContext = createContext<I18nContextValue | null>(null);
const storageKey = "gate-lang";

function initialLang(): Lang {
  const stored = localStorage.getItem(storageKey);
  if (stored === "en" || stored === "it") return stored;
  return navigator.language.toLowerCase().startsWith("it") ? "it" : "en";
}

function lookup(lang: Lang, key: string): string {
  const value = key.split(".").reduce<unknown>((node, part) => {
    if (node && typeof node === "object" && part in node) {
      return (node as Record<string, unknown>)[part];
    }
    return undefined;
  }, translations[lang]);

  return typeof value === "string" ? value : key;
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [langState, setLangState] = useState<Lang>(initialLang);

  const value = useMemo<I18nContextValue>(() => {
    function setLang(next: Lang) {
      localStorage.setItem(storageKey, next);
      setLangState(next);
    }

    return {
      lang: langState,
      setLang,
      t: (key: string) => lookup(langState, key),
    };
  }, [langState]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() {
  const context = useContext(I18nContext);
  if (!context) {
    throw new Error("useI18n must be used within I18nProvider");
  }
  return context;
}
