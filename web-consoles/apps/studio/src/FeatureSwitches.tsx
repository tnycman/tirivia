import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { switchClient, SwitchScope, type Switch } from "@tirivia/api-client";

/**
 * Back-office feature switches (ADR-0009). Chat is fully built but defaults OFF; operators
 * turn it on for testing/rehearsals or for any live show where it is wanted. Every change
 * requires a reason and is written to the audit log by the backend. Apps react within one
 * pulse (<= 5 s) via Pulse.chat_enabled / reactions_enabled.
 */
const LABELS: Record<string, { title: string; help: string }> = {
  "chat.enabled": { title: "Chat", help: "Live chat panel in the apps. Default OFF. Delivery is billed per message delivered while ON." },
  "reactions.enabled": { title: "Reactions", help: "Floating reactions and running totals." },
  "extra_lives.enabled": { title: "Extra lives", help: "Players may use an extra life (never on Q12)." },
  "signups.enabled": { title: "New sign-ups", help: "OFF pauses OTP start for new accounts." },
};

export function FeatureSwitches() {
  const [showId, setShowId] = useState("");
  const queryClient = useQueryClient();

  const switches = useQuery({
    queryKey: ["switches", showId],
    queryFn: () => switchClient.listSwitches({ showId }),
  });

  const setSwitch = useMutation({
    mutationFn: (args: { key: string; enabled: boolean; scope: SwitchScope; reason: string }) =>
      switchClient.setSwitch({
        key: args.key,
        enabled: args.enabled,
        scope: args.scope,
        showId: args.scope === SwitchScope.SHOW ? showId : "",
        reason: args.reason,
      }),
    onSettled: () => queryClient.invalidateQueries({ queryKey: ["switches"] }),
  });

  const onToggle = (sw: Switch, scope: SwitchScope) => {
    const next = !sw.effective;
    const label = LABELS[sw.key]?.title ?? sw.key;
    const where = scope === SwitchScope.SHOW ? `show ${showId}` : "ALL shows";
    // A typed reason doubles as the confirmation step; it is stored in the audit log.
    const reason = window.prompt(`Turn ${label} ${next ? "ON" : "OFF"} for ${where}?\nReason (required, audited):`);
    if (!reason?.trim()) return;
    setSwitch.mutate({ key: sw.key, enabled: next, scope, reason: reason.trim() });
  };

  return (
    <section aria-labelledby="switches-heading">
      <h2 id="switches-heading">Feature switches</h2>
      <label>
        Show id (optional, for per-show overrides):{" "}
        <input value={showId} onChange={(e) => setShowId(e.target.value.trim())} placeholder="e.g. 0193a1b2-…" />
      </label>

      {switches.isPending && <p>Loading…</p>}
      {switches.isError && <p role="alert">Could not load switches: {String(switches.error)}</p>}
      {setSwitch.isError && <p role="alert">Change rejected: {String(setSwitch.error)}</p>}

      <table style={{ width: "100%", borderCollapse: "collapse", marginTop: "1rem" }}>
        <thead>
          <tr>
            <th align="left">Switch</th>
            <th align="left">Live state</th>
            <th align="left">Source</th>
            <th align="left">Actions</th>
          </tr>
        </thead>
        <tbody>
          {(switches.data?.switches ?? []).map((sw) => (
            <tr key={`${sw.key}|${sw.showId}`} style={{ borderTop: "1px solid var(--theme-border)" }}>
              <td>
                <strong>{LABELS[sw.key]?.title ?? sw.key}</strong>
                <div style={{ fontSize: 12, opacity: 0.75 }}>{LABELS[sw.key]?.help}</div>
              </td>
              <td aria-live="polite">
                {/* Badge pairs are contrast-gated in design-tokens (onCorrect/correct 5.0:1, onWrong/wrong 8.3:1). */}
                <span
                  style={{
                    fontWeight: 700,
                    padding: "var(--space-xs) var(--space-sm)",
                    borderRadius: "var(--radius-sm)",
                    background: sw.effective ? "var(--theme-correct)" : "var(--theme-wrong)",
                    color: sw.effective ? "var(--theme-on-correct)" : "var(--theme-on-wrong)",
                  }}
                >
                  {sw.effective ? "ON" : "OFF"}
                </span>
              </td>
              <td style={{ fontSize: 12 }}>
                {sw.isDefault ? "default" : sw.scope === SwitchScope.SHOW ? `show ${sw.showId}` : "global"}
                {sw.changedBy ? ` · by ${sw.changedBy}` : ""}
              </td>
              <td>
                <button disabled={setSwitch.isPending} onClick={() => onToggle(sw, SwitchScope.GLOBAL)}>
                  Toggle globally
                </button>{" "}
                <button disabled={setSwitch.isPending || !showId} onClick={() => onToggle(sw, SwitchScope.SHOW)}>
                  Toggle for this show
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
