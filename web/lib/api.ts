import type { Player, Game, Move, LeaderboardEntry, Analysis } from "./types";

const TOKEN_KEY = "cmp_token";

function apiBase(): string {
  return process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
}

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  auth = false
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (auth) {
    const tok = getToken();
    if (tok) headers["Authorization"] = `Bearer ${tok}`;
  }
  const res = await fetch(`${apiBase()}${path}`, { ...options, headers });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? "Request failed");
  }
  return res.json() as Promise<T>;
}

export async function register(
  email: string,
  username: string,
  password: string,
  city: string
): Promise<{ token: string }> {
  return request("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, username, password, city }),
  });
}

export async function login(
  email: string,
  password: string
): Promise<{ token: string }> {
  return request("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export async function me(): Promise<Player> {
  return request("/me", {}, true);
}

export async function createGame(
  mode: string,
  color: string
): Promise<Game> {
  return request("/games", {
    method: "POST",
    body: JSON.stringify({ mode, color }),
  }, true);
}

export async function joinGame(inviteCode: string): Promise<Game> {
  return request("/games/join", {
    method: "POST",
    body: JSON.stringify({ inviteCode }),
  }, true);
}

export interface GameSnapshot {
  game: Game;
  moves: Move[];
  fen: string;
  legalMoves: string[];
  sideToMove: "white" | "black";
}

export async function getGame(id: string): Promise<GameSnapshot> {
  return request(`/games/${id}`);
}

export async function postMove(
  id: string,
  uci: string
): Promise<GameSnapshot> {
  return request(`/games/${id}/move`, {
    method: "POST",
    body: JSON.stringify({ uci }),
  }, true);
}

export async function myGames(): Promise<Game[]> {
  return request("/players/me/games", {}, true);
}

export async function leaderboard(
  city?: string,
  limit = 50
): Promise<LeaderboardEntry[]> {
  const params = new URLSearchParams({ limit: String(limit) });
  if (city) params.set("city", city);
  return request(`/leaderboard?${params}`);
}

export async function coach(gameId: string): Promise<Analysis> {
  return request(`/games/${gameId}/coach`, { method: "POST" }, true);
}

export async function upgradePro(): Promise<Player> {
  return request("/me/upgrade", { method: "POST" }, true);
}
