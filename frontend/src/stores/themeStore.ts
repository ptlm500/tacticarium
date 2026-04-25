import { create } from "zustand";
import { persist } from "zustand/middleware";

export const THEMES = ["scorpion", "spacewolf", "blood", "badmoon"] as const;
export type Theme = (typeof THEMES)[number];

export const THEME_META: Record<Theme, { label: string; hex: string; tagline: string }> = {
  scorpion: { label: "Scorpion Green", hex: "#65B345", tagline: "Ready to strike" },
  spacewolf: { label: "Space Wolf Grey", hex: "#91BFDC", tagline: "Fear denies Faith" },
  blood: { label: "Blood Red", hex: "#D2223E", tagline: "Blood for the blood god" },
  badmoon: { label: "Badmoon Yellow", hex: "#FFF200", tagline: "We'll make it Orky" },
};

interface ThemeState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme: "scorpion",
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
