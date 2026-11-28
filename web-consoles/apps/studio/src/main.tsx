import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "@trivia/ui/tokens.css";

import { FeatureSwitches } from "./FeatureSwitches";

const queryClient = new QueryClient({
  defaultOptions: { queries: { refetchInterval: 5_000, retry: 1 } },
});

const root = document.getElementById("root");
if (!root) throw new Error("missing #root");

createRoot(root).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <main style={{ maxWidth: 880, margin: "2rem auto", padding: "0 1rem" }}>
        <h1>Studio</h1>
        <FeatureSwitches />
      </main>
    </QueryClientProvider>
  </StrictMode>,
);
