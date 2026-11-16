import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from "react";
import { getUiTheme, UiTheme, UiThemeId, UI_THEMES } from "./uiThemes";

type ThemeContextValue = {
  uiThemeId: UiThemeId;
  uiTheme: UiTheme;
  themes: UiTheme[];
  setUiThemeId: (id: UiThemeId) => void;
};

const ThemeContext = createContext<ThemeContextValue | null>(null);
const storageKey = "gate-ui-theme";

function initialTheme(): UiThemeId {
  const stored = localStorage.getItem(storageKey);
  if (stored && UI_THEMES.some((t) => t.id === stored)) return stored as UiThemeId;
  return "forest";
}

function applyCssVars(theme: UiTheme) {
  const root = document.documentElement;
  Object.entries(theme.vars).forEach(([key, value]) => root.style.setProperty(key, value));
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [uiThemeId, setUiThemeIdState] = useState<UiThemeId>(initialTheme);

  useEffect(() => {
    applyCssVars(getUiTheme(uiThemeId));
  }, [uiThemeId]);

  const value = useMemo<ThemeContextValue>(() => {
    return {
      uiThemeId,
      uiTheme: getUiTheme(uiThemeId),
      themes: UI_THEMES,
      setUiThemeId: (id: UiThemeId) => {
        localStorage.setItem(storageKey, id);
        setUiThemeIdState(id);
      },
    };
  }, [uiThemeId]);

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useUiTheme() {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useUiTheme must be used within ThemeProvider");
  return ctx;
}
