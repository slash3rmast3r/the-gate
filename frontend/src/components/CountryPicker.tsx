import { useMemo, useState } from "react";
import { COUNTRIES, QUICK_BLOCK, countryLabel, parseCountryCSV, toCountryCSV } from "../lib/countries";
import { useI18n } from "../i18n/I18nProvider";

export default function CountryPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (csv: string) => void;
}) {
  const { t, lang } = useI18n();
  const [q, setQ] = useState("");
  const selected = useMemo(() => parseCountryCSV(value), [value]);

  function toggle(code: string) {
    const set = new Set(selected);
    if (set.has(code)) set.delete(code);
    else set.add(code);
    onChange(toCountryCSV([...set].sort()));
  }

  function applyQuick() {
    const set = new Set(selected);
    QUICK_BLOCK.forEach((c) => set.add(c));
    onChange(toCountryCSV([...set].sort()));
  }

  function clearAll() {
    onChange("");
  }

  const filtered = useMemo(() => {
    const needle = q.trim().toLowerCase();
    if (!needle) return COUNTRIES;
    return COUNTRIES.filter(
      (c) =>
        c.code.toLowerCase().includes(needle) ||
        c.en.toLowerCase().includes(needle) ||
        c.it.toLowerCase().includes(needle)
    );
  }, [q]);

  return (
    <div className="country-picker">
      <div className="row" style={{ marginBottom: 8 }}>
        <button type="button" className="btn" onClick={applyQuick}>
          {t("applications.quickBlock")}
        </button>
        <button type="button" className="btn" onClick={clearAll} disabled={selected.length === 0}>
          {t("applications.clearCountries")}
        </button>
      </div>
      {selected.length > 0 && (
        <div className="chip-row" style={{ marginBottom: 10 }}>
          {selected.map((code) => (
            <button key={code} type="button" className="chip on" onClick={() => toggle(code)}>
              {code} · {countryLabel(code, lang)} ×
            </button>
          ))}
        </div>
      )}
      <input
        type="search"
        placeholder={t("applications.searchCountries")}
        value={q}
        onChange={(e) => setQ(e.target.value)}
        style={{ marginBottom: 8 }}
      />
      <div className="chip-row country-grid">
        {filtered.map((c) => {
          const on = selected.includes(c.code);
          return (
            <button
              key={c.code}
              type="button"
              className={`chip ${on ? "on" : ""}`}
              onClick={() => toggle(c.code)}
              title={c.code}
            >
              {c.code} {lang === "it" ? c.it : c.en}
            </button>
          );
        })}
      </div>
      <p className="hint">{t("applications.blockedCountriesHint")}</p>
    </div>
  );
}
