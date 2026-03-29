import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useGameStore } from '../stores/gameStore';
import { useGameConnection } from '../hooks/useGameState';
import { getToken } from '../api/client';
import { factionsApi } from '../api/factions';
import { missionsApi } from '../api/missions';
import { Faction, Detachment } from '../types/faction';
import { Mission, MissionRule, Secondary } from '../types/mission';
import { ActiveSecondary } from '../types/game';
import { FactionPicker } from '../components/setup/FactionPicker';
import { DetachmentPicker } from '../components/setup/DetachmentPicker';
import { MissionPicker } from '../components/setup/MissionPicker';
import { TwistPicker } from '../components/setup/TwistPicker';
import { SecondaryModePicker } from '../components/setup/SecondaryModePicker';

const PACK_ID = 'chapter-approved-2025-26';

export function GameSetupPage() {
  const { id: gameId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { gameState } = useGameStore();

  const token = getToken();

  const { connected, sendAction } = useGameConnection(gameId!, token);

  const [factions, setFactions] = useState<Faction[]>([]);
  const [detachments, setDetachments] = useState<Detachment[]>([]);
  const [missions, setMissions] = useState<Mission[]>([]);
  const [rules, setRules] = useState<MissionRule[]>([]);
  const [secondaries, setSecondaries] = useState<Secondary[]>([]);
  const [selectedFixedIds, setSelectedFixedIds] = useState<string[]>([]);
  const [copied, setCopied] = useState(false);
  const [loadError, setLoadError] = useState('');

  useEffect(() => {
    Promise.all([
      factionsApi.list().then(setFactions),
      missionsApi.listMissions(PACK_ID).then(setMissions),
      missionsApi.listRules(PACK_ID).then(setRules),
      missionsApi.listSecondaries(PACK_ID).then(setSecondaries),
    ]).catch(() => setLoadError('Failed to load game data. Please refresh the page.'));
  }, []);

  const myPlayer = gameState?.players.find((p) => p?.userId === user?.id) ?? null;
  const opponent = gameState?.players.find((p) => p?.userId !== user?.id) ?? null;

  const fixedSecondaries = secondaries.filter((s) => s.isFixed);
  const tacticalSecondaries = secondaries.filter((s) => !s.isFixed);

  // Load detachments when faction changes
  useEffect(() => {
    if (myPlayer?.factionId) {
      factionsApi.getDetachments(myPlayer.factionId).then(setDetachments)
        .catch(() => setLoadError('Failed to load detachments'));
    } else {
      setDetachments([]);
    }
  }, [myPlayer?.factionId]);

  // Navigate to game when it starts
  useEffect(() => {
    if (gameState?.status === 'active') {
      navigate(`/game/${gameId}`);
    }
  }, [gameState?.status, gameId, navigate]);

  const handleSelectFaction = useCallback(
    (faction: Faction) => {
      sendAction('select_faction', {
        factionId: faction.id,
        factionName: faction.name,
      });
    },
    [sendAction]
  );

  const handleSelectDetachment = useCallback(
    (detachment: Detachment) => {
      sendAction('select_detachment', {
        detachmentId: detachment.id,
        detachmentName: detachment.name,
      });
    },
    [sendAction]
  );

  const handleSelectMission = useCallback(
    (mission: Mission) => {
      sendAction('select_primary_mission', {
        missionId: mission.id,
        missionName: mission.name,
      });
    },
    [sendAction]
  );

  const handleRandomMission = useCallback(() => {
    if (missions.length === 0) return;
    const m = missions[Math.floor(Math.random() * missions.length)];
    sendAction('select_primary_mission', {
      missionId: m.id,
      missionName: m.name,
    });
  }, [sendAction, missions]);

  const handleSelectTwist = useCallback(
    (rule: MissionRule) => {
      sendAction('select_twist', {
        twistId: rule.id,
        twistName: rule.name,
      });
    },
    [sendAction]
  );

  const handleRandomTwist = useCallback(() => {
    if (rules.length === 0) return;
    const r = rules[Math.floor(Math.random() * rules.length)];
    sendAction('select_twist', {
      twistId: r.id,
      twistName: r.name,
    });
  }, [sendAction, rules]);

  const handleModeChange = useCallback(
    (mode: 'fixed' | 'tactical') => {
      sendAction('select_secondary_mode', { mode });
      setSelectedFixedIds([]);
    },
    [sendAction]
  );

  const handleFixedSelect = useCallback(
    (selected: ActiveSecondary[]) => {
      const ids = selected.map((s) => s.id);
      setSelectedFixedIds(ids);
      if (selected.length === 2) {
        sendAction('set_fixed_secondaries', { secondaries: selected });
      }
    },
    [sendAction]
  );

  const handleInitDeck = useCallback(
    (deck: ActiveSecondary[]) => {
      sendAction('init_tactical_deck', { deck });
    },
    [sendAction]
  );

  const handleReady = useCallback(() => {
    sendAction('set_ready', { ready: !myPlayer?.ready });
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
        <p>{connected ? 'Loading game...' : 'Connecting...'}</p>
      </div>
    );
  }

  const hasFaction = !!myPlayer?.factionId;
  const hasDetachment = !!myPlayer?.detachmentId;
  const hasMission = !!gameState.missionId;
  const hasTwist = !!gameState.twistId;
  const hasMode = !!myPlayer?.secondaryMode;
  const hasSecondaries =
    myPlayer?.secondaryMode === 'fixed'
      ? (myPlayer?.activeSecondaries?.length ?? 0) === 2
      : (myPlayer?.tacticalDeck?.length ?? 0) > 0;
  const canReady = hasFaction && hasDetachment && hasMission && hasTwist && hasMode && hasSecondaries;

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="p-4 border-b border-gray-800">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-bold">Game Setup</h1>
          <button
            onClick={copyInviteCode}
            className="bg-gray-800 hover:bg-gray-700 border border-gray-600 px-4 py-2 rounded-lg text-sm transition-colors"
          >
            {copied ? 'Copied!' : `Invite: ${gameState.inviteCode}`}
          </button>
        </div>
        {!opponent && (
          <p className="text-yellow-400 text-sm mt-2">
            Waiting for opponent to join...
          </p>
        )}
      </header>

      <main className="max-w-md mx-auto p-6 space-y-6">
        {loadError && (
          <div className="bg-red-900/50 border border-red-700 text-red-200 px-4 py-2 rounded">
            {loadError}
          </div>
        )}

        {/* Faction Selection */}
        <section>
          <h2 className="text-lg font-semibold mb-3">Your Faction</h2>
          <FactionPicker
            factions={factions}
            selectedId={myPlayer?.factionId || ''}
            onSelect={handleSelectFaction}
          />
        </section>

        {/* Detachment Selection */}
        {hasFaction && (
          <section>
            <h2 className="text-lg font-semibold mb-3">Detachment</h2>
            <DetachmentPicker
              detachments={detachments}
              selectedId={myPlayer?.detachmentId || ''}
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
              selectedId={gameState.missionId || ''}
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
              selectedId={gameState.twistId || ''}
              onSelect={handleSelectTwist}
              onDrawRandom={handleRandomTwist}
            />
          </section>
        )}

        {/* Secondary Mission Mode */}
        {hasTwist && (
          <section>
            <h2 className="text-lg font-semibold mb-3">Secondary Missions</h2>
            <SecondaryModePicker
              mode={myPlayer?.secondaryMode || ''}
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
              {opponent.factionName || 'Selecting faction...'}
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
              ? 'bg-green-700 hover:bg-green-800 text-white'
              : 'bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white'
          }`}
        >
          {myPlayer?.ready ? 'Ready! (click to unready)' : 'Ready Up'}
        </button>
      </main>
    </div>
  );
}
