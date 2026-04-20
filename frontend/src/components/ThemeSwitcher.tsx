import { Check, Palette } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { THEMES, THEME_META, useThemeStore } from "@/stores/themeStore";

export function ThemeSwitcher() {
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const current = THEME_META[theme];

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <Palette className="size-4" />
          <span
            className="inline-block size-3 rounded-full border border-white/20"
            style={{ backgroundColor: current.hex }}
            aria-hidden
          />
          <span className="uppercase tracking-widest text-[10px]">{current.label}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-52">
        <DropdownMenuLabel className="text-[10px] uppercase tracking-widest text-muted-foreground">
          Identity Profile
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {THEMES.map((key) => {
          const meta = THEME_META[key];
          const active = key === theme;
          return (
            <DropdownMenuItem
              key={key}
              onSelect={() => setTheme(key)}
              className="flex items-center gap-3"
            >
              <span
                className="inline-block size-3 rounded-full border border-white/20"
                style={{ backgroundColor: meta.hex }}
                aria-hidden
              />
              <div className="flex-1">
                <div className="text-sm font-medium">{meta.label}</div>
                <div className="text-[10px] text-muted-foreground">{meta.tagline}</div>
              </div>
              {active && <Check className="size-4 text-primary" />}
            </DropdownMenuItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
