import { useEffect } from "react";
import { applyThemeToDocument, useThemeStore } from "@/stores/themeStore";

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const theme = useThemeStore((s) => s.theme);

  useEffect(() => {
    applyThemeToDocument(theme);
  }, [theme]);

  return <>{children}</>;
}
