import { create } from "zustand";
import { persist } from "zustand/middleware";

export const THEMES = ["tron", "ares", "clu", "athena", "aphrodite", "poseidon"] as const;
export type Theme = (typeof THEMES)[number];

export const THEME_META: Record<Theme, { label: string; hex: string; tagline: string }> = {
  tron: { label: "Tron", hex: "#00D4FF", tagline: "Programs of the Grid" },
  ares: { label: "Ares", hex: "#FF3333", tagline: "God of War" },
  clu: { label: "Clu", hex: "#FF6600", tagline: "System Administrator" },
  athena: { label: "Athena", hex: "#FFD700", tagline: "Strategic Wisdom" },
  aphrodite: { label: "Aphrodite", hex: "#FF1493", tagline: "Divine Favor" },
  poseidon: { label: "Poseidon", hex: "#0066FF", tagline: "Deep Currents" },
};

interface ThemeState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme: "tron",
      setTheme: (theme) => set({ theme }),
    }),
    {
      name: "turn-tracker-theme",
    },
  ),
);

export function applyThemeToDocument(theme: Theme) {
  if (typeof document === "undefined") return;
  document.documentElement.dataset.theme = theme;
  document.documentElement.classList.add("dark");
}
