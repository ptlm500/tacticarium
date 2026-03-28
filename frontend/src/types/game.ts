export type Phase = 'setup' | 'command' | 'movement' | 'shooting' | 'charge' | 'fight';
export type GameStatus = 'setup' | 'active' | 'completed' | 'abandoned';

export const PHASE_ORDER: Phase[] = ['command', 'movement', 'shooting', 'charge', 'fight'];

export const PHASE_LABELS: Record<Phase, string> = {
  setup: 'Setup',
  command: 'Command',
  movement: 'Movement',
  shooting: 'Shooting',
  charge: 'Charge',
  fight: 'Fight',
};

export interface SecondaryObjective {
  id: string;
  secondaryId?: string;
  customName?: string;
  customMaxVp?: number;
  vpScored: number;
}

export interface ActiveSecondary {
  id: string;
  name: string;
  description: string;
  isFixed: boolean;
  maxVp: number;
}

export interface PlayerState {
  userId: string;
  username: string;
  playerNumber: number;
  factionId: string;
  factionName: string;
  detachmentId: string;
  detachmentName: string;
  cp: number;
  vpPrimary: number;
  vpSecondary: number;
  vpGambit: number;
  vpPaint: number;
  ready: boolean;
  gambitId?: string;
  gambitDeclaredRound?: number;
  secondaries: SecondaryObjective[];
  secondaryMode: string;
  tacticalDeck: ActiveSecondary[];
  activeSecondaries: ActiveSecondary[];
  achievedSecondaries: ActiveSecondary[];
  discardedSecondaries: ActiveSecondary[];
  isChallenger: boolean;
  challengerCardId?: string;
  adaptOrDieUses: number;
}

export interface GameState {
  gameId: string;
  inviteCode: string;
  status: GameStatus;
  currentRound: number;
  currentPhase: Phase;
  activePlayer: number;
  firstTurnPlayer: number;
  missionPackId: string;
  missionId: string;
  missionName: string;
  twistId: string;
  twistName: string;
  players: [PlayerState | null, PlayerState | null];
  createdAt: string;
  completedAt?: string;
  winnerId?: string;
}

export interface GameEvent {
  eventType: string;
  playerNumber?: number;
  round?: number;
  phase?: Phase;
  data?: Record<string, unknown>;
  createdAt?: string;
}

export interface GameSummary {
  id: string;
  inviteCode: string;
  status: GameStatus;
  missionName?: string;
  createdAt: string;
  completedAt?: string;
  players: {
    userId: string;
    username: string;
    factionName?: string;
    playerNumber: number;
    totalVp: number;
  }[];
  winnerId?: string;
}
