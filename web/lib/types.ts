export interface Player {
  id: string;
  email: string;
  username: string;
  city: string;
  isPro: boolean;
  createdAt: string;
}

export interface Game {
  id: string;
  whiteId: string | null;
  blackId: string | null;
  mode: string;
  status: string;
  inviteCode?: string | null;
  pgn: string;
  result?: string | null;
}

export interface Move {
  ply: number;
  uci: string;
  san: string;
  fenAfter: string;
}

export interface LeaderboardEntry {
  username: string;
  city: string;
  elo: number;
  games: number;
  wins: number;
}

export interface Mistake {
  ply: number;
  move: string;
  severity: "inaccuracy" | "mistake" | "blunder";
  better: string;
  comment: string;
}

export interface Analysis {
  summary: string;
  mistakes: Mistake[];
}

export interface WSState {
  type: "state";
  fen: string;
  pgn: string;
  status: string;
  legalMoves: string[];
  sideToMove: "white" | "black";
}

export interface WSError {
  type: "error";
  message: string;
}

export interface WSGameOver {
  type: "gameOver";
  result: string;
}

export type WSMessage = WSState | WSError | WSGameOver;
