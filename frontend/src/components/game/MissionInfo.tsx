import { useState } from "react";

interface Props {
  missionName: string;
  twistName: string;
}

export function MissionInfo({ missionName, twistName }: Props) {
  const [expanded, setExpanded] = useState(false);

  return (
    <section>
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full bg-gray-800 hover:bg-gray-750 border border-gray-700 rounded-lg px-4 py-3 text-left flex justify-between items-center"
      >
        <span className="font-semibold">Mission Info</span>
        <span className="text-gray-400">{expanded ? "\u25B2" : "\u25BC"}</span>
      </button>
      {expanded && (
        <div className="mt-2 bg-gray-800 rounded-lg p-4 space-y-2 text-sm">
          <div>
            <span className="text-gray-400">Primary Mission: </span>
            <span className="text-white">{missionName || "None"}</span>
          </div>
          <div>
            <span className="text-gray-400">Twist: </span>
            <span className="text-white">{twistName || "None"}</span>
          </div>
        </div>
      )}
    </section>
  );
}
