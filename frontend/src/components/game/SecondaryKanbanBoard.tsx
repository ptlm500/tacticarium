import { useState } from "react";
import { ActiveSecondary } from "../../types/game";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { ACTIVE_PILE_LIMIT, Pile, filterOptions } from "./secondaryPiles";

interface Props {
  activeSecondaries: ActiveSecondary[];
  achievedSecondaries: ActiveSecondary[];
  discardedSecondaries: ActiveSecondary[];
  tacticalDeck: ActiveSecondary[];
  onMove: (secondaryId: string, fromPile: Pile, toPile: Pile, vpScored?: number) => void;
  onSelect: (s: ActiveSecondary) => void;
}

const COLUMN_ORDER: Pile[] = ["deck", "active", "discarded", "achieved"];

const COLUMN_LABELS: Record<Pile, string> = {
  deck: "Deck",
  active: "Active",
  discarded: "Discarded",
  achieved: "Achieved",
};

const COLUMN_DOT: Record<Pile, string> = {
  deck: "bg-primary",
  active: "bg-amber-400",
  discarded: "bg-muted-foreground",
  achieved: "bg-emerald-400",
};

const MOVE_TARGETS: { target: Pile; label: string }[] = [
  { target: "active", label: "→ Active" },
  { target: "deck", label: "→ Deck" },
  { target: "discarded", label: "→ Discard" },
];

export function SecondaryKanbanBoard({
  activeSecondaries,
  achievedSecondaries,
  discardedSecondaries,
  tacticalDeck,
  onMove,
  onSelect,
}: Props) {
  const [dragging, setDragging] = useState<{ cardId: string; fromPile: Pile } | null>(null);
  const [overPile, setOverPile] = useState<Pile | null>(null);
  const [achievePrompt, setAchievePrompt] = useState<{
    card: ActiveSecondary;
    fromPile: Pile;
  } | null>(null);

  const piles: Record<Pile, ActiveSecondary[]> = {
    deck: tacticalDeck,
    active: activeSecondaries,
    discarded: discardedSecondaries,
    achieved: achievedSecondaries,
  };

  function endDrag() {
    setDragging(null);
    setOverPile(null);
  }

  function handleDragStart(card: ActiveSecondary, fromPile: Pile) {
    setDragging({ cardId: card.id, fromPile });
  }

  function isDropAllowed(toPile: Pile): boolean {
    if (!dragging) return false;
    if (dragging.fromPile === toPile) return false;
    if (toPile === "active" && piles.active.length >= ACTIVE_PILE_LIMIT) return false;
    return true;
  }

  function handleColumnDrop(toPile: Pile) {
    if (!dragging) return;
    if (!isDropAllowed(toPile)) {
      endDrag();
      return;
    }
    const sourceCol = piles[dragging.fromPile];
    const card = sourceCol.find((c) => c.id === dragging.cardId);
    if (!card) {
      endDrag();
      return;
    }
    const fromPile = dragging.fromPile;

    if (toPile === "achieved") {
      const opts = filterOptions(card.scoringOptions, "tactical");
      if (opts.length > 1) {
        setAchievePrompt({ card, fromPile });
        endDrag();
        return;
      }
      const vp = opts[0]?.vp ?? card.maxVp;
      onMove(card.id, fromPile, "achieved", vp);
      endDrag();
      return;
    }

    onMove(card.id, fromPile, toPile);
    endDrag();
  }

  return (
    <div
      data-slot="secondary-kanban-board"
      className="relative overflow-hidden rounded-sm border border-primary/20 bg-card/60 backdrop-blur-sm"
    >
      {/* Scanline overlay */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 bg-[repeating-linear-gradient(0deg,transparent,transparent_2px,rgba(0,0,0,0.04)_2px,rgba(0,0,0,0.04)_4px)]"
      />

      <div className="border-b border-primary/15 px-3 py-2">
        <span className="font-mono text-[10px] uppercase tracking-widest text-foreground/50">
          Secondary Deck — Manual
        </span>
      </div>

      <div className="grid gap-2 bg-primary/5 p-2 md:grid-cols-4">
        {COLUMN_ORDER.map((pile) => {
          const cards = piles[pile];
          const dropAllowed = dragging ? isDropAllowed(pile) : true;
          const isHover = overPile === pile && !!dragging;
          const atCapacity = pile === "active" && cards.length >= ACTIVE_PILE_LIMIT;

          return (
            <div
              key={pile}
              onDragOver={(e) => {
                if (!dragging) return;
                if (!dropAllowed) {
                  e.dataTransfer.dropEffect = "none";
                  return;
                }
                e.preventDefault();
                e.dataTransfer.dropEffect = "move";
                if (overPile !== pile) setOverPile(pile);
              }}
              onDragLeave={() => {
                if (overPile === pile) setOverPile(null);
              }}
              onDrop={(e) => {
                e.preventDefault();
                handleColumnDrop(pile);
              }}
              className={cn(
                "flex min-h-[160px] flex-col rounded-sm border transition-colors",
                isHover && dropAllowed
                  ? "border-primary/60 bg-primary/10"
                  : isHover && !dropAllowed
                    ? "border-destructive/50 bg-destructive/10"
                    : "border-border/40 bg-background/40",
              )}
              aria-label={`${COLUMN_LABELS[pile]} pile`}
            >
              <div className="flex items-center gap-2 border-b border-border/30 px-2 py-1.5">
                <span className={cn("size-1.5 rounded-full", COLUMN_DOT[pile])} />
                <span className="font-mono text-[10px] uppercase tracking-widest text-foreground/60">
                  {COLUMN_LABELS[pile]}
                </span>
                <span
                  className={cn(
                    "ml-auto rounded-sm border px-1.5 py-0.5 font-mono text-[9px]",
                    atCapacity
                      ? "border-amber-400/50 bg-amber-400/10 text-amber-300"
                      : "border-border/40 bg-background/40 text-muted-foreground",
                  )}
                >
                  {pile === "active" ? `${cards.length} / ${ACTIVE_PILE_LIMIT}` : cards.length}
                </span>
              </div>

              <div className="flex flex-1 flex-col gap-1.5 p-1.5">
                {cards.length === 0 ? (
                  <div
                    className={cn(
                      "flex flex-1 items-center justify-center rounded-sm border border-dashed text-[10px] uppercase tracking-widest text-muted-foreground/60",
                      isHover && dropAllowed ? "border-primary/40" : "border-border/30",
                    )}
                  >
                    {dragging && dropAllowed ? "drop here" : "empty"}
                  </div>
                ) : (
                  cards.map((card) => (
                    <SecondaryCard
                      key={`${card.id}-${pile}`}
                      card={card}
                      pile={pile}
                      dragging={dragging?.cardId === card.id}
                      onDragStart={() => handleDragStart(card, pile)}
                      onDragEnd={endDrag}
                      onSelect={() => onSelect(card)}
                      onMove={onMove}
                      activeAtCapacity={piles.active.length >= ACTIVE_PILE_LIMIT}
                    />
                  ))
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* Corner ticks */}
      <div className="pointer-events-none absolute left-0 top-0 size-2.5 border-l-2 border-t-2 border-primary/40" />
      <div className="pointer-events-none absolute right-0 top-0 size-2.5 border-r-2 border-t-2 border-primary/40" />
      <div className="pointer-events-none absolute bottom-0 left-0 size-2.5 border-b-2 border-l-2 border-primary/40" />
      <div className="pointer-events-none absolute bottom-0 right-0 size-2.5 border-b-2 border-r-2 border-primary/40" />

      <AchievePromptDialog
        prompt={achievePrompt}
        onCancel={() => setAchievePrompt(null)}
        onPick={(vp) => {
          if (!achievePrompt) return;
          onMove(achievePrompt.card.id, achievePrompt.fromPile, "achieved", vp);
          setAchievePrompt(null);
        }}
      />
    </div>
  );
}

function SecondaryCard({
  card,
  pile,
  dragging,
  onDragStart,
  onDragEnd,
  onSelect,
  onMove,
  activeAtCapacity,
}: {
  card: ActiveSecondary;
  pile: Pile;
  dragging: boolean;
  onDragStart: () => void;
  onDragEnd: () => void;
  onSelect: () => void;
  onMove: Props["onMove"];
  activeAtCapacity: boolean;
}) {
  const scoringOpts = filterOptions(card.scoringOptions, "tactical");
  const showAchieveButtons = pile !== "achieved" && scoringOpts.length > 0;

  return (
    <div
      draggable
      onDragStart={(e) => {
        onDragStart();
        e.dataTransfer.effectAllowed = "move";
        if (e.currentTarget instanceof HTMLElement) {
          e.dataTransfer.setDragImage(e.currentTarget, 12, 12);
        }
      }}
      onDragEnd={onDragEnd}
      className={cn(
        "group relative cursor-grab rounded-sm border border-border/60 bg-background/70 p-2 transition-all active:cursor-grabbing",
        "hover:border-primary/40 hover:shadow-[0_0_8px_rgba(120,180,255,0.08)]",
        dragging && "scale-[0.98] opacity-40",
        pile === "achieved" && "border-emerald-500/40 bg-emerald-500/5",
      )}
    >
      {/* Drag-handle dots */}
      <div className="pointer-events-none absolute left-1 top-1/2 flex -translate-y-1/2 flex-col gap-[2px] opacity-30 transition-opacity group-hover:opacity-70">
        <span className="size-[2px] rounded-full bg-foreground" />
        <span className="size-[2px] rounded-full bg-foreground" />
        <span className="size-[2px] rounded-full bg-foreground" />
      </div>

      <button
        type="button"
        onClick={onSelect}
        className="block w-full pl-3 text-left"
        title="View full details"
      >
        <div className="flex items-start justify-between gap-2">
          <span
            className={cn(
              "text-xs font-medium",
              pile === "achieved" ? "text-emerald-300" : "text-foreground",
            )}
          >
            {card.name}
          </span>
          <span className="font-mono text-[9px] uppercase tracking-widest text-muted-foreground">
            {card.maxVp} VP
          </span>
        </div>
        {card.description && pile === "active" && (
          <p className="mt-1 line-clamp-2 text-[11px] text-muted-foreground">{card.description}</p>
        )}
        {card.drawRestriction && (
          <span className="mt-1 inline-block rounded-sm border border-amber-500/40 bg-amber-500/10 px-1 py-0.5 font-mono text-[8px] uppercase tracking-widest text-amber-300">
            R{card.drawRestriction.round} {card.drawRestriction.mode}
          </span>
        )}
      </button>

      {/* Move buttons (kept as accessible / touch fallback) */}
      <div className="mt-2 flex flex-wrap gap-1 pl-3">
        {MOVE_TARGETS.filter(({ target }) => target !== pile).map(({ target, label }) => {
          const disabled = target === "active" && activeAtCapacity;
          return (
            <Button
              key={target}
              type="button"
              size="sm"
              variant="outline"
              className="h-6 px-2 text-[10px]"
              disabled={disabled}
              onClick={() => onMove(card.id, pile, target)}
              title={disabled ? "Active pile is full" : undefined}
            >
              {label}
            </Button>
          );
        })}
        {showAchieveButtons &&
          scoringOpts.map((opt, i) => (
            <Button
              key={i}
              type="button"
              size="sm"
              onClick={() => onMove(card.id, pile, "achieved", opt.vp)}
              title={opt.label}
              className="h-6 bg-emerald-600 px-2 text-[10px] text-white hover:bg-emerald-700"
            >
              ✓ {opt.label} +{opt.vp}
            </Button>
          ))}
      </div>
    </div>
  );
}

function AchievePromptDialog({
  prompt,
  onPick,
  onCancel,
}: {
  prompt: { card: ActiveSecondary; fromPile: Pile } | null;
  onPick: (vp: number) => void;
  onCancel: () => void;
}) {
  const opts = filterOptions(prompt?.card.scoringOptions, "tactical");
  return (
    <Dialog
      open={prompt !== null}
      onOpenChange={(next) => {
        if (!next) onCancel();
      }}
    >
      {prompt && (
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle className="font-mono uppercase tracking-widest text-primary">
              Achieve: {prompt.card.name}
            </DialogTitle>
            <DialogDescription className="font-mono text-[10px] uppercase tracking-widest">
              Pick the scoring outcome
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-2">
            {opts.map((opt, i) => (
              <Button
                key={i}
                type="button"
                onClick={() => onPick(opt.vp)}
                className="justify-between bg-emerald-600 text-white hover:bg-emerald-700"
              >
                <span>{opt.label}</span>
                <span className="font-mono">+{opt.vp} VP</span>
              </Button>
            ))}
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
          </div>
        </DialogContent>
      )}
    </Dialog>
  );
}
