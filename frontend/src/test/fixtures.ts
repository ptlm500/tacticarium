import { GameState, PlayerState, GameEvent, SecondaryCard, Board, Objective } from "../types/game";
import { Faction, Detachment, Stratagem } from "../types/faction";
import { Mission, MissionCard, ForceDisposition, DeploymentPattern } from "../types/mission";
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
    vpPaint: 0,
    ready: false,
    secondaryMode: "tactical",
    fixedSecondaryIds: [],
    secondaryDeck: [],
    secondaryHand: [],
    secondaryScored: [],
    pendingScorePrompts: [],
    primaryScoredThisRound: 0,
    secondaryScoredThisRound: 0,
    missionId: "mission-1",
    missionName: "Supply Drop",
    cpGainedThisRound: 0,
    stratagemsUsedThisPhase: [],
    ...overrides,
  };
}

const mockBoard: Board = {
  deploymentPatternId: "",
  objectives: [],
  playerSides: [],
};

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
    board: mockBoard,
    vpPerGameCap: 50,
    vpPerRoundCap: 15,
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

export const mockActiveSecondary: SecondaryCard = {
  id: "sec-1",
  name: "Behind Enemy Lines",
  text: "Score VP for units in enemy deployment zone",
  card_type: "secondary",
  awards: [],
};

export const mockFixedSecondary: SecondaryCard = {
  id: "sec-fixed-1",
  name: "Assassination",
  text: "Score VP for destroying enemy characters",
  card_type: "secondary",
  awards: [],
};

export const mockFactions: Faction[] = [
  { id: "faction-sm", name: "Space Marines" },
  { id: "faction-csm", name: "Chaos Space Marines" },
  { id: "faction-orks", name: "Orks" },
];

export const mockDetachments: Detachment[] = [
  {
    id: "det-gladius",
    factionId: "faction-sm",
    name: "Gladius Task Force",
    detachmentPoints: 3,
    forceDispositions: ["take-and-hold"],
  },
  {
    id: "det-ironstorm",
    factionId: "faction-sm",
    name: "Ironstorm Spearhead",
    detachmentPoints: 2,
    forceDispositions: ["hold-the-line"],
  },
];

export const mockStratagems: Stratagem[] = [
  {
    id: "strat-1",
    factionId: "faction-sm",
    name: "Command Re-roll",
    type: "Core",
    cpCost: 1,
    playerTurn: "Either player's turn",
    phases: ["Any phase"],
  },
  {
    id: "strat-2",
    factionId: "faction-sm",
    detachmentId: "det-gladius",
    name: "Storm of Fire",
    type: "Battle Tactic",
    cpCost: 1,
    playerTurn: "Your turn",
    phases: ["Shooting phase"],
  },
  {
    id: "strat-3",
    factionId: "faction-sm",
    name: "Heroic Intervention",
    type: "Strategic Ploy",
    cpCost: 2,
    playerTurn: "Opponent's turn",
    phases: ["Charge phase"],
  },
  {
    id: "strat-challenger",
    factionId: "faction-sm",
    name: "Banner of Defiance",
    type: "Challenger \u2013 Battle Tactic Stratagem",
    cpCost: 1,
    playerTurn: "Your turn",
    phases: ["Any phase"],
  },
];

export const mockMissions: Mission[] = [
  {
    id: "mission-1",
    name: "Supply Drop",
    vpPerGameCap: 50,
    vpPerRoundCap: 15,
    deploymentPatternIds: [],
  },
  {
    id: "mission-2",
    name: "Scorched Earth",
    vpPerGameCap: 50,
    vpPerRoundCap: 15,
    deploymentPatternIds: [],
  },
];

export const mockForceDispositions: ForceDisposition[] = [
  { id: "take-and-hold", name: "Take and Hold", text: "Press forward and seize the centre." },
  { id: "hold-the-line", name: "Hold the Line", text: "Defend your territory at all costs." },
];

export function makeObjective(overrides?: Partial<Objective>): Objective {
  return {
    index: 0,
    point: { x: 30, y: 22 },
    role: "central",
    controlledBy: 0,
    tags: [],
    ...overrides,
  };
}

export const mockObjectives: Objective[] = [
  makeObjective({ index: 0, point: { x: 30, y: 22 }, role: "central" }),
  makeObjective({
    index: 1,
    point: { x: 22, y: 8 },
    role: "home",
    homeSide: "defender",
    controlledBy: 1,
  }),
  makeObjective({ index: 2, point: { x: 46, y: 10 }, role: "expansion", controlledBy: 2 }),
];

export const mockBoardWithObjectives: Board = {
  deploymentPatternId: "tipping-point",
  objectives: mockObjectives,
  playerSides: ["defender", "attacker"],
};

export const mockDeploymentPatterns: DeploymentPattern[] = [
  {
    id: "tipping-point",
    name: "Tipping Point",
    objectives: [
      { x: 30, y: 22 },
      { x: 22, y: 8 },
      { x: 46, y: 10 },
    ],
    territories: [
      {
        player: "defender",
        shape: {
          type: "polygon",
          points: [
            { x: 0, y: 0 },
            { x: 30, y: 0 },
            { x: 30, y: 44 },
            { x: 0, y: 44 },
          ],
        },
        position: { x: 0, y: 0 },
      },
    ],
    zones: [
      {
        player: "defender",
        name: "Defender Deployment",
        shape: {
          type: "polygon",
          points: [
            { x: 0, y: 0 },
            { x: 12, y: 0 },
            { x: 12, y: 44 },
            { x: 0, y: 44 },
          ],
        },
        position: { x: 0, y: 0 },
        color: "#3b82f6",
      },
    ],
    recommendedTerrainLayoutIds: [],
  },
];

export const mockSecondaryCards: MissionCard[] = [
  {
    id: "sec-behind-lines",
    name: "Behind Enemy Lines",
    cardType: "secondary",
    text: "Score VP for units in enemy deployment zone.",
  },
  {
    id: "sec-assassination",
    name: "Assassination",
    cardType: "secondary",
    text: "Score VP for destroying enemy characters.",
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
