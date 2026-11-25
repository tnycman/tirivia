// Independent second implementation of the vector runner, using only Node built-ins
// (node:crypto). Proves the corpus is language-neutral: nothing here shares code with the Go
// reference. Covers shuffle order, envelope parsing, Ed25519 verification, key lookup, size and
// prefix rules. CueBody field checks stay in the typed runners (Go/Swift/Kotlin).
//
//   node protocol/testvectors/runners/node/verify-vectors.mjs
import { createHmac, createPublicKey, verify } from "node:crypto";
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const dir = join(dirname(fileURLToPath(import.meta.url)), "..", "..");
const load = (f) => JSON.parse(readFileSync(join(dir, f), "utf8"));
let failures = 0;
const fail = (msg) => { failures++; console.error("FAIL", msg); };

// ---------------------------------------------------------------- shuffle
function order(seed, questionId, n) {
  const k = createHmac("sha256", seed).update(Buffer.from(questionId, "utf8")).digest(); // Uint8 (unsigned)
  const idx = [...Array(n).keys()];
  for (let i = n - 1; i >= 1; i--) {
    const j = k[n - 1 - i] % (i + 1);
    [idx[i], idx[j]] = [idx[j], idx[i]];
  }
  return idx;
}
const sh = load("shuffle_vectors.json");
for (const v of sh.vectors) {
  const got = order(Buffer.from(v.seed_hex, "hex"), v.question_id, v.n);
  if (JSON.stringify(got) !== JSON.stringify(v.order)) fail(`shuffle ${v.name}: got ${got} want ${v.order}`);
}
console.log(`shuffle: ${sh.vectors.length} vectors checked`);

// ---------------------------------------------------------------- cue envelope + signature
function readVarint(buf, pos) {
  let result = 0n, shift = 0n;
  for (;;) {
    if (pos >= buf.length) throw new Error("truncated varint");
    const b = buf[pos++];
    result |= BigInt(b & 0x7f) << shift;
    if (!(b & 0x80)) return [result, pos];
    shift += 7n;
    if (shift > 63n) throw new Error("varint too long");
  }
}
// Minimal protobuf parse of SignedCue {1: bytes body, 2: bytes signature, 3: uint32 key_id}.
function parseSignedCue(buf) {
  const out = { body: null, signature: null, keyId: 0n };
  let pos = 0;
  while (pos < buf.length) {
    let tag; [tag, pos] = readVarint(buf, pos);
    const field = Number(tag >> 3n), wire = Number(tag & 7n);
    if (wire === 2) {
      let len; [len, pos] = readVarint(buf, pos);
      const end = pos + Number(len);
      if (end > buf.length) throw new Error("truncated field");
      const val = buf.subarray(pos, end); pos = end;
      if (field === 1) out.body = val; else if (field === 2) out.signature = val;
    } else if (wire === 0) {
      let val; [val, pos] = readVarint(buf, pos);
      if (field === 3) out.keyId = val;
    } else throw new Error("unexpected wire type " + wire);
  }
  return out;
}
const ed25519Key = (raw) =>
  createPublicKey({ key: Buffer.concat([Buffer.from("302a300506032b6570032100", "hex"), raw]), format: "der", type: "spki" });

const cs = load("cue_vectors.json");
const keys = new Map(Object.entries(cs.public_keys).map(([id, b64]) => [BigInt(id), ed25519Key(Buffer.from(b64, "base64"))]));
const B64URL = /^[A-Za-z0-9_-]*$/;

function envelopeOutcome(text) {
  if (Buffer.byteLength(text, "utf8") > cs.max_text_bytes) return "reject:too_large";
  if (!text.startsWith("t1.")) return "reject:malformed";
  const enc = text.slice(3);
  if (!B64URL.test(enc) || enc.length % 4 === 1) return "reject:malformed";
  let env;
  try { env = parseSignedCue(Buffer.from(enc, "base64url")); } catch { return "reject:malformed"; }
  if (!env.body || !env.signature) return "reject:malformed";
  const key = keys.get(env.keyId);
  if (!key) return "reject:unknown_key";
  if (!verify(null, env.body, key, env.signature)) return "reject:bad_signature";
  return "signature_ok";
}
let checked = 0;
for (const v of cs.vectors) {
  const got = envelopeOutcome(v.text);
  const expected = v.expect.split("|");
  // Envelope-level runner: valid / wrong_show / replay / gap all require a good signature.
  const sigLevel = expected.every((e) => !e.startsWith("reject:") || ["reject:wrong_show", "reject:replay"].includes(e));
  const ok = sigLevel ? got === "signature_ok" : expected.includes(got);
  if (!ok) fail(`cue ${v.name}: got ${got}, want ${v.expect}`);
  checked++;
}
console.log(`cue: ${checked} vectors checked (envelope, size, key, Ed25519)`);
if (failures) { console.error(`${failures} failure(s)`); process.exit(1); }
console.log("all vectors agree with an independent implementation");
