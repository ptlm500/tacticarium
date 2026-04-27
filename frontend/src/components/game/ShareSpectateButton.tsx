import { useState } from "react";
import { Check, Eye } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";

interface Props {
  gameId: string;
  variant?: "outline" | "ghost";
  size?: "sm" | "icon";
  className?: string;
}

export function ShareSpectateButton({
  gameId,
  variant = "outline",
  size = "sm",
  className,
}: Props) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    const url = `${window.location.origin}/game/${gameId}/spectate`;
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Could not copy spectate link");
    }
  };

  if (size === "icon") {
    return (
      <Button
        type="button"
        variant={variant}
        size="icon"
        onClick={() => void copy()}
        aria-label="Copy spectate link"
        title="Copy spectate link"
        className={className}
      >
        {copied ? <Check className="size-3.5" /> : <Eye className="size-3.5" />}
      </Button>
    );
  }

  return (
    <Button
      type="button"
      variant={variant}
      size="sm"
      onClick={() => void copy()}
      className={`gap-2 font-mono uppercase tracking-widest ${className ?? ""}`}
    >
      {copied ? (
        <>
          <Check className="size-3.5" />
          Copied!
        </>
      ) : (
        <>
          <Eye className="size-3.5" />
          Spectate Link
        </>
      )}
    </Button>
  );
}
