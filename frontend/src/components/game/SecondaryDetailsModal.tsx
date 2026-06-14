import { SecondaryCard } from "../../types/game";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";

interface Props {
  secondary: SecondaryCard | null;
  onClose: () => void;
}

export function SecondaryDetailsModal({ secondary, onClose }: Props) {
  return (
    <Dialog
      open={secondary !== null}
      onOpenChange={(next) => {
        if (!next) onClose();
      }}
    >
      {secondary && (
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="font-mono uppercase tracking-widest text-primary">
              {secondary.name}
            </DialogTitle>
          </DialogHeader>

          <p className="whitespace-pre-line text-sm text-foreground/90">{secondary.text}</p>
        </DialogContent>
      )}
    </Dialog>
  );
}
