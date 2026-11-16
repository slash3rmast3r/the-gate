export type Country = { code: string; en: string; it: string };

/** ISO 3166-1 alpha-2 list used by the geo-block picker. */
export const COUNTRIES: Country[] = [
  { code: "AF", en: "Afghanistan", it: "Afghanistan" },
  { code: "AL", en: "Albania", it: "Albania" },
  { code: "DZ", en: "Algeria", it: "Algeria" },
  { code: "AR", en: "Argentina", it: "Argentina" },
  { code: "AM", en: "Armenia", it: "Armenia" },
  { code: "AU", en: "Australia", it: "Australia" },
  { code: "AT", en: "Austria", it: "Austria" },
  { code: "AZ", en: "Azerbaijan", it: "Azerbaigian" },
  { code: "BH", en: "Bahrain", it: "Bahrain" },
  { code: "BD", en: "Bangladesh", it: "Bangladesh" },
  { code: "BY", en: "Belarus", it: "Bielorussia" },
  { code: "BE", en: "Belgium", it: "Belgio" },
  { code: "BO", en: "Bolivia", it: "Bolivia" },
  { code: "BA", en: "Bosnia and Herzegovina", it: "Bosnia ed Erzegovina" },
  { code: "BR", en: "Brazil", it: "Brasile" },
  { code: "BG", en: "Bulgaria", it: "Bulgaria" },
  { code: "KH", en: "Cambodia", it: "Cambogia" },
  { code: "CM", en: "Cameroon", it: "Camerun" },
  { code: "CA", en: "Canada", it: "Canada" },
  { code: "CL", en: "Chile", it: "Cile" },
  { code: "CN", en: "China", it: "Cina" },
  { code: "CO", en: "Colombia", it: "Colombia" },
  { code: "CD", en: "Congo (DRC)", it: "Congo (RDC)" },
  { code: "CR", en: "Costa Rica", it: "Costa Rica" },
  { code: "HR", en: "Croatia", it: "Croazia" },
  { code: "CU", en: "Cuba", it: "Cuba" },
  { code: "CY", en: "Cyprus", it: "Cipro" },
  { code: "CZ", en: "Czechia", it: "Cechia" },
  { code: "DK", en: "Denmark", it: "Danimarca" },
  { code: "DO", en: "Dominican Republic", it: "Repubblica Dominicana" },
  { code: "EC", en: "Ecuador", it: "Ecuador" },
  { code: "EG", en: "Egypt", it: "Egitto" },
  { code: "EE", en: "Estonia", it: "Estonia" },
  { code: "ET", en: "Ethiopia", it: "Etiopia" },
  { code: "FI", en: "Finland", it: "Finlandia" },
  { code: "FR", en: "France", it: "Francia" },
  { code: "GE", en: "Georgia", it: "Georgia" },
  { code: "DE", en: "Germany", it: "Germania" },
  { code: "GH", en: "Ghana", it: "Ghana" },
  { code: "GR", en: "Greece", it: "Grecia" },
  { code: "HK", en: "Hong Kong", it: "Hong Kong" },
  { code: "HU", en: "Hungary", it: "Ungheria" },
  { code: "IS", en: "Iceland", it: "Islanda" },
  { code: "IN", en: "India", it: "India" },
  { code: "ID", en: "Indonesia", it: "Indonesia" },
  { code: "IR", en: "Iran", it: "Iran" },
  { code: "IQ", en: "Iraq", it: "Iraq" },
  { code: "IE", en: "Ireland", it: "Irlanda" },
  { code: "IL", en: "Israel", it: "Israele" },
  { code: "IT", en: "Italy", it: "Italia" },
  { code: "JP", en: "Japan", it: "Giappone" },
  { code: "JO", en: "Jordan", it: "Giordania" },
  { code: "KZ", en: "Kazakhstan", it: "Kazakistan" },
  { code: "KE", en: "Kenya", it: "Kenya" },
  { code: "KP", en: "North Korea", it: "Corea del Nord" },
  { code: "KR", en: "South Korea", it: "Corea del Sud" },
  { code: "KW", en: "Kuwait", it: "Kuwait" },
  { code: "LV", en: "Latvia", it: "Lettonia" },
  { code: "LB", en: "Lebanon", it: "Libano" },
  { code: "LY", en: "Libya", it: "Libia" },
  { code: "LT", en: "Lithuania", it: "Lituania" },
  { code: "LU", en: "Luxembourg", it: "Lussemburgo" },
  { code: "MO", en: "Macao", it: "Macao" },
  { code: "MY", en: "Malaysia", it: "Malaysia" },
  { code: "MT", en: "Malta", it: "Malta" },
  { code: "MX", en: "Mexico", it: "Messico" },
  { code: "MD", en: "Moldova", it: "Moldavia" },
  { code: "MA", en: "Morocco", it: "Marocco" },
  { code: "MM", en: "Myanmar", it: "Myanmar" },
  { code: "NP", en: "Nepal", it: "Nepal" },
  { code: "NL", en: "Netherlands", it: "Paesi Bassi" },
  { code: "NZ", en: "New Zealand", it: "Nuova Zelanda" },
  { code: "NG", en: "Nigeria", it: "Nigeria" },
  { code: "MK", en: "North Macedonia", it: "Macedonia del Nord" },
  { code: "NO", en: "Norway", it: "Norvegia" },
  { code: "OM", en: "Oman", it: "Oman" },
  { code: "PK", en: "Pakistan", it: "Pakistan" },
  { code: "PS", en: "Palestine", it: "Palestina" },
  { code: "PA", en: "Panama", it: "Panama" },
  { code: "PY", en: "Paraguay", it: "Paraguay" },
  { code: "PE", en: "Peru", it: "Perù" },
  { code: "PH", en: "Philippines", it: "Filippine" },
  { code: "PL", en: "Poland", it: "Polonia" },
  { code: "PT", en: "Portugal", it: "Portogallo" },
  { code: "QA", en: "Qatar", it: "Qatar" },
  { code: "RO", en: "Romania", it: "Romania" },
  { code: "RU", en: "Russia", it: "Russia" },
  { code: "SA", en: "Saudi Arabia", it: "Arabia Saudita" },
  { code: "RS", en: "Serbia", it: "Serbia" },
  { code: "SG", en: "Singapore", it: "Singapore" },
  { code: "SK", en: "Slovakia", it: "Slovacchia" },
  { code: "SI", en: "Slovenia", it: "Slovenia" },
  { code: "ZA", en: "South Africa", it: "Sudafrica" },
  { code: "ES", en: "Spain", it: "Spagna" },
  { code: "LK", en: "Sri Lanka", it: "Sri Lanka" },
  { code: "SE", en: "Sweden", it: "Svezia" },
  { code: "CH", en: "Switzerland", it: "Svizzera" },
  { code: "SY", en: "Syria", it: "Siria" },
  { code: "TW", en: "Taiwan", it: "Taiwan" },
  { code: "TZ", en: "Tanzania", it: "Tanzania" },
  { code: "TH", en: "Thailand", it: "Thailandia" },
  { code: "TR", en: "Türkiye", it: "Turchia" },
  { code: "UA", en: "Ukraine", it: "Ucraina" },
  { code: "AE", en: "United Arab Emirates", it: "Emirati Arabi Uniti" },
  { code: "GB", en: "United Kingdom", it: "Regno Unito" },
  { code: "US", en: "United States", it: "Stati Uniti" },
  { code: "UY", en: "Uruguay", it: "Uruguay" },
  { code: "UZ", en: "Uzbekistan", it: "Uzbekistan" },
  { code: "VE", en: "Venezuela", it: "Venezuela" },
  { code: "VN", en: "Vietnam", it: "Vietnam" },
  { code: "YE", en: "Yemen", it: "Yemen" },
];

export const QUICK_BLOCK = ["CN", "RU", "KP", "IR", "BY"];

export function parseCountryCSV(raw: string): string[] {
  const parts = (raw || "").toUpperCase().split(/[,;\s]+/).map((p) => p.trim()).filter(Boolean);
  const out: string[] = [];
  const seen = new Set<string>();
  for (const p of parts) {
    if (p.length !== 2 || seen.has(p)) continue;
    seen.add(p);
    out.push(p);
  }
  return out;
}

export function toCountryCSV(codes: string[]): string {
  return codes.map((c) => c.toUpperCase()).filter((c) => c.length === 2).join(",");
}

export function countryLabel(code: string, lang: "en" | "it"): string {
  const c = COUNTRIES.find((x) => x.code === code.toUpperCase());
  if (!c) return code.toUpperCase();
  return lang === "it" ? c.it : c.en;
}
