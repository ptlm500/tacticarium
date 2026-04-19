import { X } from "lucide-react";
import { AlertBanner } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

interface ErrorBannerProps {
  message: string;
  onDismiss?: () => void;
}

export function ErrorBanner({ message, onDismiss }: ErrorBannerProps) {
  return (
    <div className="relative">
      <AlertBanner variant="danger" subtitle="Error" title={message} />
      {onDismiss && (
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={onDismiss}
          aria-label="Dismiss"
          className="absolute right-2 top-1/2 -translate-y-1/2 text-destructive hover:bg-destructive/20"
        >
          <X className="size-4" />
        </Button>
      )}
    </div>
  );
}
