import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// In production the SPA is served by the BFF, which sets a strict nonce-based CSP
// (strict-dynamic, object-src 'none', base-uri 'none', frame-ancestors 'none').
export default defineConfig({
  plugins: [react()],
  server: {
    // Local development: proxy /api to the BFF (or, in dev mode only, the admin listener).
    proxy: { "/api": { target: "http://localhost:19090", rewrite: (p) => p.replace(/^\/api/, "") } },
  },
  build: { sourcemap: false },
});
