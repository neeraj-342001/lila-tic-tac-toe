/** Opcodes must match `backend/match.go` */
export const Op = {
  Start: 1,
  Update: 2,
  Done: 3,
  Move: 4,
  Rejected: 5,
  OpponentLeft: 6,
} as const;

export type Board = number[];

export interface StartPayload {
  board: Board;
  marks: Record<string, number>;
  mark: number;
  deadline: number;
}

export interface UpdatePayload {
  board: Board;
  mark: number;
  deadline: number;
}

export interface DonePayload {
  board: Board;
  winner: number;
  winner_positions?: number[];
  next_game_start: number;
  timeout?: boolean;
}
