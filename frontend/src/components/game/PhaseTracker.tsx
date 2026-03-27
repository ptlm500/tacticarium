import { Phase, PHASE_LABELS } from '../../types/game';

interface Props {
  currentPhase: Phase;
  phases: Phase[];
}

export function PhaseTracker({ currentPhase, phases }: Props) {
  return (
    <div className="flex items-center gap-1 overflow-x-auto">
      {phases.map((phase, i) => {
        const isActive = phase === currentPhase;
        const isPast = phases.indexOf(currentPhase) > i;
        return (
          <div key={phase} className="flex items-center">
            {i > 0 && (
              <div
                className={`w-4 h-0.5 ${isPast ? 'bg-indigo-500' : 'bg-gray-700'}`}
              />
            )}
            <span
              className={`text-xs px-2 py-1 rounded whitespace-nowrap ${
                isActive
                  ? 'bg-indigo-600 text-white font-semibold'
                  : isPast
                  ? 'text-indigo-400'
                  : 'text-gray-500'
              }`}
            >
              {PHASE_LABELS[phase]}
            </span>
          </div>
        );
      })}
    </div>
  );
}
