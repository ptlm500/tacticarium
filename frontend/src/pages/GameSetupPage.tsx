import { useEffect, useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { useGameStore } from "../stores/gameStore";
import { useGameConnection } from "../hooks/useGameState";
import { getToken } from "../api/client";
import { Faction, Detachment } from "../types/faction";
import { Mission, MissionRule } from "../types/mission";
import { ActiveSecondary } from "../types/game";
import { FactionPicker } from "../components/setup/FactionPicker";
import { DetachmentPicker } from "../components/setup/DetachmentPicker";
import { MissionPicker } from "../components/setup/MissionPicker";
import { TwistPicker } from "../components/setup/TwistPicker";
import { FirstPlayerPicker } from "../components/setup/FirstPlayerPicker";
import { SecondaryModePicker } from "../components/setup/SecondaryModePicker";
import { useFactions, useDetachments } from "../hooks/queries/useFactionQueries";
import { useMissions, useMissionRules, useSecondaries } from "../hooks/queries/useMissionQueries";

const PACK_ID = "chapter-approved-2025-26";

export function GameSetupPage() {
  const { id: gameId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { gameState } = useGameStore();

  const token = getToken();

  const { connected, sendAction } = useGameConnection(gameId!, token);

  const { data: factions = [] } = useFactions();
  const { data: missions = [] } = useMissions(PACK_ID);
  const { data: rules = [] } = useMissionRules(PACK_ID);
  const { data: secondaries = [] } = useSecondaries(PACK_ID);

  const [selectedFixedIds, setSelectedFixedIds] = useState<string[]>([]);
  const [copied, setCopied] = useState(false);

  const myPlayer = gameState?.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState?.players.find((p) => p?.userId !== user?.id) ?? null;

  const { data: detachments = [] } = useDetachments(myPlayer?.factionId);

  const fixedSecondaries = secondaries.filter((s) => s.isFixed);
  const tacticalSecondaries = secondaries.filter((s) => !s.isFixed);

  // Navigate to game when it starts
  useEffect(() => {
    if (gameState?.status === "active") {
      navigate(`/game/${gameId}`);
    }
  }, [gameState?.status, gameId, navigate]);

  const handleSelectFaction = useCallback(
    (faction: Faction) => {
      sendAction("select_faction", {
        factionId: faction.id,
        factionName: faction.name,
      });
    },
    [sendAction],
  );

  const handleSelectDetachment = useCallback(
    (detachment: Detachment) => {
      sendAction("select_detachment", {
        detachmentId: detachment.id,
        detachmentName: detachment.name,
      });
    },
    [sendAction],
  );

  const handleSelectMission = useCallback(
    (mission: Mission) => {
      sendAction("select_primary_mission", {
        missionPackId: PACK_ID,
        missionId: mission.id,
        missionName: mission.name,
      });
    },
    [sendAction],
  );

  const handleRandomMission = useCallback(() => {
    if (missions.length === 0) return;
    const m = missions[Math.floor(Math.random() * missions.length)];
    sendAction("select_primary_mission", {
      missionPackId: PACK_ID,
      missionId: m.id,
      missionName: m.name,
    });
  }, [sendAction, missions]);

  const handleSelectTwist = useCallback(
    (rule: MissionRule) => {
      sendAction("select_twist", {
        twistId: rule.id,
        twistName: rule.name,
      });
    },
    [sendAction],
  );

  const handleRandomTwist = useCallback(() => {
    if (rules.length === 0) return;
    const r = rules[Math.floor(Math.random() * rules.length)];
    sendAction("select_twist", {
      twistId: r.id,
      twistName: r.name,
    });
  }, [sendAction, rules]);

  const handleSelectFirstPlayer = useCallback(
    (playerNumber: 1 | 2) => {
      sendAction("select_first_turn_player", { playerNumber });
    },
    [sendAction],
  );

  const handleRandomFirstPlayer = useCallback(() => {
    const playerNumber = Math.random() < 0.5 ? 1 : 2;
    sendAction("select_first_turn_player", { playerNumber });
  }, [sendAction]);

  const handleModeChange = useCallback(
    (mode: "fixed" | "tactical") => {
      sendAction("select_secondary_mode", { mode });
      setSelectedFixedIds([]);
    },
    [sendAction],
  );

  const handleFixedSelect = useCallback(
    (selected: ActiveSecondary[]) => {
      const ids = selected.map((s) => s.id);
      setSelectedFixedIds(ids);
      if (selected.length === 2) {
        sendAction("set_fixed_secondaries", { secondaries: selected });
      }
    },
    [sendAction],
  );

  const handleInitDeck = useCallback(
    (deck: ActiveSecondary[]) => {
      sendAction("init_tactical_deck", { deck });
    },
    [sendAction],
  );

  const handleReady = useCallback(() => {
    sendAction("set_ready", { ready: !myPlayer?.ready });
  }, [sendAction, myPlayer?.ready]);

  const copyInviteCode = () => {
    if (gameState?.inviteCode) {
      navigator.clipboard.writeText(gameState.inviteCode);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (!gameState) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <p>{connected ? "Loading game..." : "Connecting..."}</p>
      </div>
    );
  }

  const hasFaction = !!myPlayer?.factionId;
  const hasDetachment = !!myPlayer?.detachmentId;
  const hasMission = !!gameState.missionId;
  const hasTwist = !!gameState.twistId;
  const hasFirstPlayer = (gameState.firstTurnPlayer ?? 0) > 0;
  const hasMode = !!myPlayer?.secondaryMode;
  const hasSecondaries =
    myPlayer?.secondaryMode === "fixed"
      ? (myPlayer?.activeSecondaries?.length ?? 0) === 2
      : (myPlayer?.tacticalDeck?.length ?? 0) > 0;
  const canReady =
    hasFaction &&
    hasDetachment &&
    hasMission &&
    hasTwist &&
    hasFirstPlayer &&
    hasMode &&
    hasSecondaries;

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="p-4 border-b border-gray-800">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold">Game Setup</h1>
          <button
            onClick={copyInviteCode}
            className="bg-gray-800 hover:bg-gray-700 border border-gray-600 px-4 py-2 rounded-lg text-sm transition-colors"
          >
            {copied ? "Copied!" : `Invite: ${gameState.inviteCode}`}
          </button>
        </div>
        {!opponent && (
          <p className="text-yellow-400 text-sm mt-2">Waiting for opponent to join...</p>
        )}
      </header>

      <main className="max-w-md mx-auto p-6 space-y-6">
        {/* Faction Selection */}
        <section>
          <h2 className="text-lg font-semibold mb-3">Your Faction</h2>
          <FactionPicker
            factions={factions}
            selectedId={myPlayer?.factionId || ""}
            onSelect={handleSelectFaction}
          />
        </section>

        {/* Detachment Selection */}
        {hasFaction && (
          <section>
            <h2 className="text-lg font-semibold mb-3">Detachment</h2>
            <DetachmentPicker
              detachments={detachments}
              selectedId={myPlayer?.detachmentId || ""}
              onSelect={handleSelectDetachment}
            />
          </section>
        )}

        {/* Primary Mission */}
        {hasDetachment && (
          <section>
            <h2 className="text-lg font-semibold mb-3">Primary Mission</h2>
            <MissionPicker
              missions={missions}
              selectedId={gameState.missionId || ""}
              onSelect={handleSelectMission}
              onDrawRandom={handleRandomMission}
            />
          </section>
        )}

        {/* Twist / Mission Rule */}
        {hasMission && (
          <section>
            <h2 className="text-lg font-semibold mb-3">Twist</h2>
            <TwistPicker
              rules={rules}
              selectedId={gameState.twistId || ""}
              onSelect={handleSelectTwist}
              onDrawRandom={handleRandomTwist}
            />
          </section>
        )}

        {/* First Turn Player */}
        {hasTwist && myPlayer && (
          <section>
            <h2 className="text-lg font-semibold mb-3">Who Goes First?</h2>
            <FirstPlayerPicker
              myPlayerNumber={myPlayer.playerNumber}
              myUsername={myPlayer.username}
              opponentUsername={opponent?.username}
              selected={gameState.firstTurnPlayer ?? 0}
              onSelect={handleSelectFirstPlayer}
              onRandom={handleRandomFirstPlayer}
            />
          </section>
        )}

        {/* Secondary Mission Mode */}
        {hasFirstPlayer && (
          <section>
            <h2 className="text-lg font-semibold mb-3">Secondary Missions</h2>
            <SecondaryModePicker
              mode={myPlayer?.secondaryMode || ""}
              onModeChange={handleModeChange}
              fixedSecondaries={fixedSecondaries}
              selectedFixedIds={selectedFixedIds}
              onFixedSelect={handleFixedSelect}
              tacticalSecondaries={tacticalSecondaries}
              deckInitialized={(myPlayer?.tacticalDeck?.length ?? 0) > 0}
              onInitDeck={handleInitDeck}
            />
          </section>
        )}

        {/* Opponent Status */}
        {opponent && (
          <section className="bg-gray-800 rounded-lg p-4">
            <h2 className="text-sm font-semibold text-gray-400 mb-2">
              Opponent: {opponent.username}
            </h2>
            <p className="text-sm">
              {opponent.factionName || "Selecting faction..."}
              {opponent.detachmentName && ` - ${opponent.detachmentName}`}
            </p>
            <p className="text-sm mt-1">
              {opponent.ready ? (
                <span className="text-green-400">Ready</span>
              ) : (
                <span className="text-yellow-400">Not ready</span>
              )}
            </p>
          </section>
        )}

        {/* Ready Button */}
        <button
          onClick={handleReady}
          disabled={!canReady}
          className={`w-full font-semibold py-3 rounded-lg transition-colors ${
            myPlayer?.ready
              ? "bg-green-700 hover:bg-green-800 text-white"
              : "bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white"
          }`}
        >
          {myPlayer?.ready ? "Ready! (click to unready)" : "Ready Up"}
        </button>
      </main>
    </div>
  );
}
