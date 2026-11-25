# web-consoles: back office and public website

React 19.3 and TypeScript 7 single-page apps, built with Vite 8. They sit behind a **backend-for-frontend (BFF)** and do not use React Server Components. The public site uses **Astro 7**. Stack details are in [`stack/08-web-consoles`](../stack/08-web-consoles/README.md).

```
web-consoles/
├── apps/
│   ├── studio/          # run of show, question release, bursts, FEATURE SWITCHES (chat etc.)  ← built
│   ├── question-bank/   # authoring + two-person approval + content limits (240 B / 64 B)
│   ├── claims/          # claim review, four-eyes payout release
│   ├── risk/            # clear/review/block queue
│   ├── clips/           # clip review and publishing
│   └── website/         # Astro static site: sq (default), en, mk
├── packages/
│   ├── api-client/      # generated from protocol/ (protobuf-es 2.17) + Connect-ES 2.3 clients  ← built
│   ├── ui  auth (BFF session, step-up)  approvals (four-eyes)  audit  i18n  config  test-utils
└── bff/                 # SSO (FIDO2-only), __Host- session cookies, strict CSP, forwards to internal admin listener
```

## Feature switches (back-office chat toggle)

`apps/studio/src/FeatureSwitches.tsx` lists every switch and shows its **live effective state**. Each switch can be toggled:

- **globally**, or
- **for a single show**, which overrides the global value.

Every change requires a reason, and the backend writes it to the audit log. Chat is **off by default**. Turn it on for testing and rehearsals, or for any live show where it is wanted ([ADR-0009](../docs/adr/0009-chat-backend-switch.md)).

## Commands

```sh
npm install
npm run generate      # regenerate the TS client from protocol/proto (needs buf on PATH)
npm run typecheck
npm run build
npm run dev -w @tirivia/studio   # proxies /api to the admin listener (ADMIN_DEV_MODE=1 locally only)
```

**Verified on 2026-11-25:**

- `npm install` succeeded.
- The client generated with `protoc-gen-es` 2.17.0.
- `tsc` (TypeScript 7.0.2) passed with no errors.
- `vite build` (8.3.8) produced the studio bundle.
