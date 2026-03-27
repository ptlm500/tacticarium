import { GameEvent, GameState } from './game';

export type ServerMessageType =
  | 'state_update'
  | 'event'
  | 'error'
  | 'pong'
  | 'player_connected'
  | 'player_disconnected';

export interface ServerMessage {
  type: ServerMessageType;
  data: unknown;
}

export interface StateUpdateMessage {
  type: 'state_update';
  data: GameState;
}

export interface EventMessage {
  type: 'event';
  data: GameEvent;
}

export interface ErrorMessage {
  type: 'error';
  data: { message: string; code: string };
}

export interface PlayerConnectionMessage {
  type: 'player_connected' | 'player_disconnected';
  data: { playerNumber: number; username?: string };
}

export type ClientMessageType = 'action' | 'ping' | 'sync_request';

export interface ClientMessage {
  type: ClientMessageType;
  data?: {
    type: string;
    [key: string]: unknown;
  };
}
