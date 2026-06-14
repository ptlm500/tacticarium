import { Shuffle } from "lucide-react";
import { MissionCard } from "../../types/mission";
import { SecondaryCard } from "../../types/game";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface Props {
  mode: string;
  onModeChange: (mode: "fixed" | "tactical") => void;
  fixedSecondaries: MissionCard[];
  selectedFixedIds: string[];
  onFixedSelect: (secondaries: SecondaryCard[]) => void;
  tacticalSecondaries: MissionCard[];
  deckInitialized: boolean;
  onInitDeck: (deck: SecondaryCard[]) => void;
}

function shuffleArray<T>(arr: T[]): T[] {
  const shuffled = [...arr];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

function secondaryToCard(s: MissionCard): SecondaryCard {
  return {
    id: s.id,
    name: s.name,
    text: s.text ?? "",
    card_type: "secondary",
    awards: [],
  };
}

export function SecondaryModePicker({
  mode,
  onModeChange,
  fixedSecondaries,
  selectedFixedIds,
  onFixedSelect,
  tacticalSecondaries,
  deckInitialized,
  onInitDeck,
}: Props) {
  const handleFixedToggle = (secondary: MissionCard) => {
    const isSelected = selectedFixedIds.includes(secondary.id);
    let newIds: string[];
    if (isSelected) {
      newIds = selectedFixedIds.filter((id) => id !== secondary.id);
    } else {
      if (selectedFixedIds.length >= 2) return;
      newIds = [...selectedFixedIds, secondary.id];
    }
    const selected = fixedSecondaries.filter((s) => newIds.includes(s.id)).map(secondaryToCard);
    onFixedSelect(selected);
  };

  const handleInitDeck = () => {
    const deck = shuffleArray(tacticalSecondaries.map(secondaryToCard));
    onInitDeck(deck);
  };

  const modeBtn = (active: boolean) =>
    cn(
      "flex-1 rounded-sm border px-4 py-2 text-sm font-mono uppercase tracking-widest transition-colors",
      active
        ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
        : "border-border/60 bg-background/40 text-muted-foreground hover:border-primary/50 hover:text-foreground",
    );

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        <button
          type="button"
          onClick={() => onModeChange("fixed")}
          className={modeBtn(mode === "fixed")}
        >
          Fixed
        </button>
        <button
          type="button"
          onClick={() => onModeChange("tactical")}
          className={modeBtn(mode === "tactical")}
        >
          Tactical
        </button>
      </div>

      {mode === "fixed" && (
        <div className="space-y-2">
          <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
            Select exactly 2 fixed secondary missions ({selectedFixedIds.length}/2)
          </p>
          <div className="max-h-60 space-y-2 overflow-y-auto pr-1">
            {fixedSecondaries.map((s) => {
              const active = selectedFixedIds.includes(s.id);
              return (
                <button
                  key={s.id}
                  type="button"
                  onClick={() => handleFixedToggle(s)}
                  className={cn(
                    "flex w-full items-center justify-between gap-2 rounded-sm border p-3 text-left text-sm transition-colors",
                    active
                      ? "border-primary bg-primary/10 text-primary shadow-[0_0_8px_var(--primary)]"
                      : "border-border/60 bg-background/40 text-foreground hover:border-primary/50 hover:bg-primary/5",
                  )}
                >
                  <span className="font-medium">{s.name}</span>
                </button>
              );
            })}
          </div>
        </div>
      )}

      {mode === "tactical" && (
        <div className="space-y-3">
          <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
            Deck of {tacticalSecondaries.length} tactical missions · Draw 2 each command phase
          </p>
          {!deckInitialized ? (
            <Button
              type="button"
              onClick={handleInitDeck}
              className="w-full gap-2 font-mono uppercase tracking-widest"
            >
              <Shuffle className="size-4" />
              Shuffle &amp; Initialize Deck
            </Button>
          ) : (
            <p className="font-mono text-[10px] uppercase tracking-widest text-emerald-400">
              Deck initialized and ready
            </p>
          )}
        </div>
      )}
    </div>
  );
}
