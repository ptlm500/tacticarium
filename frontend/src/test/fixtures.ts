import { GameState, PlayerState, GameEvent, ActiveSecondary } from "../types/game";
import { Faction, Detachment, Stratagem } from "../types/faction";
import { Mission, Secondary, MissionRule } from "../types/mission";
import { User } from "../api/auth";

export const mockUser: User = {
  id: "user-1",
  username: "TestPlayer",
  createdAt: "2025-01-01T00:00:00Z",
};

export const mockOpponentUser: User = {
  id: "user-2",
  username: "Opponent",
  createdAt: "2025-01-01T00:00:00Z",
};

export function makePlayerState(overrides?: Partial<PlayerState>): PlayerState {
  return {
    userId: "user-1",
    username: "TestPlayer",
    playerNumber: 1,
    factionId: "faction-sm",
    factionName: "Space Marines",
    detachmentId: "det-gladius",
    detachmentName: "Gladius Task Force",
    cp: 5,
    vpPrimary: 0,
    vpSecondary: 0,
    vpGambit: 0,
    vpPaint: 0,
    ready: false,
    secondaries: [],
    secondaryMode: "tactical",
    tacticalDeck: [],
    activeSecondaries: [],
    achievedSecondaries: [],
    discardedSecondaries: [],
    cpGainedThisRound: 0,
    isChallenger: false,
    adaptOrDieUses: 0,
    stratagemsUsedThisPhase: [],
    newOrdersUsedThisPhase: false,
    vpPrimaryScoredSlots: {},
    ...overrides,
  };
}

export function makeGameState(overrides?: Partial<GameState>): GameState {
  return {
    gameId: "game-1",
    inviteCode: "ABC123",
    status: "active",
    currentRound: 1,
    currentTurn: 1,
    currentPhase: "command",
    activePlayer: 1,
    firstTurnPlayer: 1,
    missionPackId: "chapter-approved-2025-26",
    missionId: "mission-1",
    missionName: "Supply Drop",
    twistId: "twist-1",
    twistName: "Hidden Supplies",
    players: [
      makePlayerState(),
      makePlayerState({
        userId: "user-2",
        username: "Opponent",
        playerNumber: 2,
        factionId: "faction-csm",
        factionName: "Chaos Space Marines",
        detachmentId: "det-black-legion",
        detachmentName: "Black Legion",
      }),
    ],
    createdAt: "2025-01-01T00:00:00Z",
    ...overrides,
  };
}

export const mockActiveSecondary: ActiveSecondary = {
  id: "sec-1",
  name: "Behind Enemy Lines",
  description: "Score VP for units in enemy deployment zone",
  isFixed: false,
  maxVp: 5,
  scoringOptions: [
    { label: "1 unit", vp: 2 },
    { label: "2+ units", vp: 5, mode: "tactical" },
    { label: "End game", vp: 4, mode: "fixed" },
  ],
};

export const mockFixedSecondary: ActiveSecondary = {
  id: "sec-fixed-1",
  name: "Assassination",
  description: "Score VP for destroying enemy characters",
  isFixed: true,
  maxVp: 8,
  scoringOptions: [
    { label: "Character destroyed", vp: 3, mode: "fixed" },
    { label: "Warlord destroyed", vp: 5, mode: "fixed" },
  ],
};

export const mockFactions: Faction[] = [
  { id: "faction-sm", name: "Space Marines" },
  { id: "faction-csm", name: "Chaos Space Marines" },
  { id: "faction-orks", name: "Orks" },
];

export const mockDetachments: Detachment[] = [
  { id: "det-gladius", factionId: "faction-sm", name: "Gladius Task Force" },
  { id: "det-ironstorm", factionId: "faction-sm", name: "Ironstorm Spearhead" },
];

export const mockStratagems: Stratagem[] = [
  {
    id: "strat-1",
    factionId: "faction-sm",
    name: "Command Re-roll",
    type: "Core",
    cpCost: 1,
    turn: "Either player's turn",
    phase: "Any phase",
    description: "Re-roll one hit roll, wound roll, or saving throw.",
  },
  {
    id: "strat-2",
    factionId: "faction-sm",
    detachmentId: "det-gladius",
    name: "Storm of Fire",
    type: "Battle Tactic",
    cpCost: 1,
    turn: "Your turn",
    phase: "Shooting phase",
    description: "Improve AP of ranged weapons by 1.",
  },
  {
    id: "strat-3",
    factionId: "faction-sm",
    name: "Heroic Intervention",
    type: "Strategic Ploy",
    cpCost: 2,
    turn: "Opponent's turn",
    phase: "Charge phase",
    description: "Heroically intervene with a character.",
  },
  {
    id: "strat-challenger",
    factionId: "faction-sm",
    name: "Banner of Defiance",
    type: "Challenger \u2013 Battle Tactic Stratagem",
    cpCost: 1,
    turn: "Your turn",
    phase: "Any phase",
    description: "Challenger stratagem that should not appear in the general panel.",
  },
];

export const mockMissions: Mission[] = [
  {
    id: "mission-1",
    missionPackId: "chapter-approved-2025-26",
    name: "Supply Drop",
    lore: "Secure the supply crates.",
    description: "Control objectives to score VP.",
    scoringRules: [
      { label: "2 objectives", vp: 5, minRound: 2, scoringTiming: "end_of_command_phase" },
      { label: "3+ objectives", vp: 10, minRound: 2, scoringTiming: "end_of_command_phase" },
    ],
    scoringTiming: "end_of_command_phase",
  },
  {
    id: "mission-2",
    missionPackId: "chapter-approved-2025-26",
    name: "Scorched Earth",
    lore: "Burn it all.",
    description: "Destroy objectives to score VP.",
    scoringRules: [
      { label: "Burned 1", vp: 4 },
      { label: "Burned 2+", vp: 8 },
    ],
    scoringTiming: "end_of_command_phase",
  },
];

export const mockRules: MissionRule[] = [
  {
    id: "twist-1",
    missionPackId: "chapter-approved-2025-26",
    name: "Hidden Supplies",
    lore: "Supplies are scattered.",
    description: "Additional objectives appear.",
  },
  {
    id: "twist-2",
    missionPackId: "chapter-approved-2025-26",
    name: "Chilling Rain",
    lore: "Rain falls.",
    description: "Reduce visibility.",
  },
];

export const mockSecondaries: Secondary[] = [
  {
    id: "sec-behind-lines",
    missionPackId: "chapter-approved-2025-26",
    name: "Behind Enemy Lines",
    lore: "Infiltrate.",
    description: "Score VP for units in enemy deployment zone.",
    maxVp: 5,
    isFixed: false,
    scoringOptions: [
      { label: "1 unit", vp: 2 },
      { label: "2+ units", vp: 5, mode: "tactical" },
    ],
  },
  {
    id: "sec-assassination",
    missionPackId: "chapter-approved-2025-26",
    name: "Assassination",
    lore: "Kill their leaders.",
    description: "Score VP for destroying enemy characters.",
    maxVp: 8,
    isFixed: true,
    scoringOptions: [
      { label: "Character", vp: 3, mode: "fixed" },
      { label: "Warlord", vp: 5, mode: "fixed" },
    ],
  },
];

export const mockEvent: GameEvent = {
  eventType: "phase_advanced",
  playerNumber: 1,
  round: 1,
  phase: "command",
  data: {},
  createdAt: "2025-01-01T00:01:00Z",
};
