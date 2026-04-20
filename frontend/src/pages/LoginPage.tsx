import { LogIn } from "lucide-react";
import { Button } from "@/components/ui/button";
import { HUDFrame } from "@/components/ui/hud-frame";
import { Separator } from "@/components/ui/separator";
import { ThemeSwitcher } from "@/components/ThemeSwitcher";
import { useAuth } from "../hooks/useAuth";

export function LoginPage() {
  const { login } = useAuth();

  return (
    <div className="relative min-h-screen overflow-hidden bg-background">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-[0.04]"
        style={{
          backgroundImage:
            "linear-gradient(var(--primary) 1px, transparent 1px), linear-gradient(90deg, var(--primary) 1px, transparent 1px)",
          backgroundSize: "40px 40px",
        }}
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_center,transparent_0%,var(--background)_75%)]"
      />

      <div className="absolute right-4 top-4 z-10">
        <ThemeSwitcher />
      </div>

      <main className="relative z-0 flex min-h-screen items-center justify-center px-4">
        <HUDFrame label="Identity Required" className="w-full max-w-md">
          <div className="space-y-6 py-4 text-center">
            <div className="space-y-1">
              <div className="font-mono text-[24px] uppercase tracking-[0.3em] text-primary">
                Tacticarium
              </div>
              <p className="text-sm text-muted-foreground">
                Warhammer 40K 10th Edition · Turn Tracker
              </p>
            </div>

            <Separator />

            <Button
              onClick={login}
              size="lg"
              className="w-full gap-2 font-mono tracking-widest uppercase"
            >
              <LogIn className="size-4" />
              Sign in with Discord
            </Button>

            <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground/60">
              Authenticate to deploy
            </p>
          </div>
        </HUDFrame>
      </main>
    </div>
  );
}
