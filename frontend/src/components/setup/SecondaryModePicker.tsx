import { Secondary } from "../../types/mission";
import { ActiveSecondary } from "../../types/game";

interface Props {
  mode: string;
  onModeChange: (mode: "fixed" | "tactical") => void;
  fixedSecondaries: Secondary[];
  selectedFixedIds: string[];
  onFixedSelect: (secondaries: ActiveSecondary[]) => void;
  tacticalSecondaries: Secondary[];
  deckInitialized: boolean;
  onInitDeck: (deck: ActiveSecondary[]) => void;
}

function shuffleArray<T>(arr: T[]): T[] {
  const shuffled = [...arr];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

function secondaryToActive(s: Secondary): ActiveSecondary {
  return {
    id: s.id,
    name: s.name,
    description: s.description,
    isFixed: s.isFixed,
    maxVp: s.maxVp,
    scoringOptions: s.scoringOptions,
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
  const handleFixedToggle = (secondary: Secondary) => {
    const isSelected = selectedFixedIds.includes(secondary.id);
    let newIds: string[];
    if (isSelected) {
      newIds = selectedFixedIds.filter((id) => id !== secondary.id);
    } else {
      if (selectedFixedIds.length >= 2) return;
      newIds = [...selectedFixedIds, secondary.id];
    }
    const selected = fixedSecondaries.filter((s) => newIds.includes(s.id)).map(secondaryToActive);
    onFixedSelect(selected);
  };

  const handleInitDeck = () => {
    const deck = shuffleArray(tacticalSecondaries.map(secondaryToActive));
    onInitDeck(deck);
  };

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        <button
          onClick={() => onModeChange("fixed")}
          className={`flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-colors ${
            mode === "fixed"
              ? "bg-indigo-600 text-white"
              : "bg-gray-800 text-gray-400 hover:bg-gray-750 border border-gray-700"
          }`}
        >
          Fixed
        </button>
        <button
          onClick={() => onModeChange("tactical")}
          className={`flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-colors ${
            mode === "tactical"
              ? "bg-indigo-600 text-white"
              : "bg-gray-800 text-gray-400 hover:bg-gray-750 border border-gray-700"
          }`}
        >
          Tactical
        </button>
      </div>

      {mode === "fixed" && (
        <div className="space-y-2">
          <p className="text-sm text-gray-400">
            Select exactly 2 fixed secondary missions ({selectedFixedIds.length}/2)
          </p>
          <div className="space-y-2 max-h-60 overflow-y-auto">
            {fixedSecondaries.map((s) => {
              const isSelected = selectedFixedIds.includes(s.id);
              return (
                <button
                  key={s.id}
                  onClick={() => handleFixedToggle(s)}
                  className={`w-full p-3 rounded-lg text-left text-sm transition-colors ${
                    isSelected
                      ? "bg-indigo-600 text-white border-2 border-indigo-400"
                      : "bg-gray-800 hover:bg-gray-750 border border-gray-700 text-gray-300"
                  }`}
                >
                  <span className="font-medium">{s.name}</span>
                  <span className="text-xs ml-2 opacity-70">({s.maxVp} VP)</span>
                </button>
              );
            })}
          </div>
        </div>
      )}

      {mode === "tactical" && (
        <div className="space-y-3">
          <p className="text-sm text-gray-400">
            Your deck of {tacticalSecondaries.length} tactical secondary missions will be shuffled.
            You'll draw 2 at the start of each command phase.
          </p>
          {!deckInitialized ? (
            <button
              onClick={handleInitDeck}
              className="w-full bg-indigo-600 hover:bg-indigo-700 py-2 rounded-lg text-sm font-medium transition-colors"
            >
              Shuffle & Initialize Deck
            </button>
          ) : (
            <p className="text-sm text-green-400">Deck initialized and ready.</p>
          )}
        </div>
      )}
    </div>
  );
}
