import { createClient, type Client } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { SwitchService } from "./gen/tirivia/admin/v1/switch_service_pb";

export * from "./gen/tirivia/admin/v1/switch_service_pb";

/**
 * Consoles never call the backend directly: every request goes to the same-origin BFF
 * (`/api`), which holds the operator's server-side session (__Host- cookie, FIDO2 SSO)
 * and forwards to the internal admin listener. No tokens live in the browser.
 */
const transport = createConnectTransport({
  baseUrl: "/api",
  useBinaryFormat: true,
  fetch: (input, init) => fetch(input, { ...init, credentials: "same-origin" }),
});

export type SwitchClient = Client<typeof SwitchService>;

export const switchClient: SwitchClient = createClient(SwitchService, transport);
