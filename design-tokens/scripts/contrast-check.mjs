// WCAG 2.2 contrast gate for every text/background pairing the apps use.
// Formula: https://www.w3.org/TR/WCAG22/#dfn-relative-luminance and #dfn-contrast-ratio
// Thresholds: SC 1.4.3 (AA) 4.5:1 normal text, 3:1 large text (>= 24 px, or >= 18.66 px bold).
// Overlay text on live video is checked against the WORST case: a pure-white frame behind the scrim.
import { readFileSync, readdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const tokensDir = join(dirname(fileURLToPath(import.meta.url)), "..", "tokens");
const load = (f) => JSON.parse(readFileSync(join(tokensDir, f), "utf8"));

function flatten(obj, prefix = [], out = {}) {
  for (const [k, v] of Object.entries(obj)) {
    if (k.startsWith("$")) continue;
    if (v && typeof v === "object" && "$value" in v) out[[...prefix, k].join(".")] = v.$value;
    else if (v && typeof v === "object") flatten(v, [...prefix, k], out);
  }
  return out;
}
const resolve = (tokens, value, depth = 0) => {
  if (depth > 10) throw new Error("alias loop");
  if (typeof value === "string" && /^\{.+\}$/.test(value)) {
    const target = tokens[value.slice(1, -1)];
    if (target === undefined) throw new Error("unresolved alias " + value);
    return resolve(tokens, target, depth + 1);
  }
  return value;
};

const lin = (c) => (c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4);
const luminance = ([r, g, b]) => 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
const ratio = (a, b) => {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
};
// Source-over compositing in sRGB space (how browsers/OS compositors blend).
const over = (fg, alpha, bg) => fg.map((c, i) => c * alpha + bg[i] * (1 - alpha));

const primitives = load("primitives.tokens.json");
const overlay = load("overlay.tokens.json");
const themes = readdirSync(tokensDir).filter((f) => f.startsWith("theme-"));
let failures = 0;
const rows = [];
const check = (label, fg, bg, min) => {
  const r = ratio(fg, bg);
  const ok = r >= min;
  if (!ok) failures++;
  rows.push(`${ok ? "PASS" : "FAIL"}  ${r.toFixed(2).padStart(5)}:1  (min ${min})  ${label}`);
};

const NORMAL = 4.5, LARGE = 3.0;
for (const file of themes) {
  const t = flatten({ ...primitives, ...load(file) });
  const c = (name) => resolve(t, t[name]).components;
  const th = file.replace(".tokens.json", "");
  for (const [fg, bg, min] of [
    ["theme.text", "theme.bg", NORMAL], ["theme.text", "theme.surface", NORMAL],
    ["theme.textMuted", "theme.bg", NORMAL], ["theme.textMuted", "theme.surface", NORMAL],
    ["theme.onPrimary", "theme.primary", NORMAL], ["theme.primaryText", "theme.bg", NORMAL], ["theme.primaryText", "theme.surface", NORMAL],
    ["theme.onCorrect", "theme.correct", NORMAL], ["theme.onWrong", "theme.wrong", NORMAL],
    ["theme.warningText", "theme.bg", LARGE],
  ]) check(`${th}: ${fg} on ${bg}`, c(fg), c(bg), min);
}

{
  const t = flatten({ ...primitives, ...overlay });
  const scrim = resolve(t, t["overlay.scrim"]);
  const worstFrame = [1, 1, 1]; // pure white video frame behind the scrim
  const bg = over(scrim.components, scrim.alpha, worstFrame);
  const c = (name) => resolve(t, t[name]).components;
  for (const [fg, min] of [
    ["overlay.text", NORMAL], ["overlay.textMuted", NORMAL], ["overlay.correct", NORMAL],
    ["overlay.wrong", NORMAL], ["overlay.timerWarning", LARGE],
  ]) check(`overlay: ${fg} on scrim over white video`, c(fg), bg, min);
}

console.log(rows.join("\n"));
if (failures) {
  console.error(`\n${failures} pairing(s) below WCAG 2.2 AA`);
  process.exit(1);
}
console.log(`\nall ${rows.length} pairings meet WCAG 2.2 AA`);
