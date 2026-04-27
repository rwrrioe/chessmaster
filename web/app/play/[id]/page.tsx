"use client";
import { useEffect, useState, useCallback, useRef } from "react";
import { useParams } from "next/navigation";
import type { Game, Move } from "@/lib/types";
import type { WSState } from "@/lib/types";
import { getGame, postMove, getToken } from "@/lib/api";
import { connectGameSocket, WSController } from "@/lib/ws";
import Board from "@/components/Board";
import MoveList from "@/components/MoveList";
import CoachPanel from "@/components/CoachPanel";
import Bezel from "@/components/Bezel";
import EyebrowTag from "@/components/EyebrowTag";
import Reveal from "@/components/Reveal";
import { Flag, ChartLineUp } from "phosphor-react";
import { cn } from "@/lib/cn";

const INITIAL_FEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    active: "Game in progress",
    pending: "Waiting for opponent",
    white_won: "White wins",
    black_won: "Black wins",
    draw: "Draw",
    Ongoing: "Game in progress",
    Check: "Check!",
    Checkmate: "Checkmate",
    Stalemate: "Stalemate",
  };
  return map[status] ?? status;
}

function isTerminal(status: string): boolean {
  return ["white_won", "black_won", "draw", "Checkmate", "Stalemate"].includes(status);
}

export default function GamePage() {
  const params = useParams();
  const gameId = params.id as string;

  const [game, setGame] = useState<Game | null>(null);
  const [moves, setMoves] = useState<Move[]>([]);
  const [wsState, setWsState] = useState<WSState | null>(null);
  const [wsConnected, setWsConnected] = useState(false);
  const [showCoach, setShowCoach] = useState(false);
  const [loading, setLoading] = useState(true);
  const wsRef = useRef<WSController | null>(null);

  // Derive FEN and legalMoves from WS state or initial
  const fen = wsState?.fen ?? INITIAL_FEN;
  const legalMoves = wsState?.legalMoves ?? [];
  const status = wsState?.status ?? game?.status ?? "pending";
  const sideToMove = wsState?.sideToMove ?? "white";

  // Fetch initial game state
  useEffect(() => {
    async function load() {
      try {
        const data = await getGame(gameId);
        setGame(data.game);
        setMoves(data.moves);
      } catch {
        // Ignore — we'll show skeleton
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [gameId]);

  // Connect WebSocket
  useEffect(() => {
    const token = getToken();
    if (!token) return; // spectator mode — REST only

    const controller = connectGameSocket(gameId, token, {
      onState: (msg) => {
        setWsState(msg);
        setWsConnected(true);
      },
      onGameOver: (msg) => {
        // Refresh game data
        getGame(gameId).then((d) => {
          setGame(d.game);
          setMoves(d.moves);
        });
      },
      onError: () => {
        // WS error — will fall back to REST for moves
      },
      onClose: () => setWsConnected(false),
    });
    wsRef.current = controller;
    return () => controller.close();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gameId]);

  const handleMove = useCallback(
    async (uci: string) => {
      if (wsConnected && wsRef.current) {
        wsRef.current.send({ type: "move", uci });
      } else {
        // REST fallback
        try {
          const data = await postMove(gameId, uci);
          setGame(data.game);
          setMoves(data.moves);
        } catch {
          // Ignore invalid move
        }
      }
    },
    [wsConnected, gameId]
  );

  const handleResign = useCallback(() => {
    if (wsConnected && wsRef.current) {
      wsRef.current.send({ type: "resign" });
    }
  }, [wsConnected]);

  if (loading) {
    return (
      <main className="min-h-[100dvh] flex items-center justify-center">
        <span className="w-8 h-8 rounded-full border border-white/20 border-t-white/70 animate-spin" />
      </main>
    );
  }

  return (
    <main className="min-h-[100dvh] px-4 pt-28 pb-16 max-w-6xl mx-auto">
      <Reveal>
        <div className="flex items-center gap-3 mb-8">
          <EyebrowTag>{game?.mode ?? "game"}</EyebrowTag>
          <span
            className={cn(
              "text-xs rounded-full px-3 py-1 ring-1",
              isTerminal(status)
                ? "ring-white/20 text-white/60 bg-white/5"
                : "ring-emerald-500/30 text-emerald-400/80 bg-emerald-500/10"
            )}
          >
            {statusLabel(status)}
          </span>
          {wsConnected && (
            <span className="text-[10px] text-emerald-400/50 uppercase tracking-widest">
              Live
            </span>
          )}
        </div>
      </Reveal>

      <div className="grid grid-cols-1 lg:grid-cols-[1fr_320px] gap-6">
        {/* Board */}
        <Reveal delay={60}>
          <Board
            fen={fen}
            legalMoves={legalMoves}
            onMove={handleMove}
            disabled={isTerminal(status)}
          />
        </Reveal>

        {/* Sidebar */}
        <Reveal delay={120} className="space-y-4">
          {/* Side indicator */}
          <Bezel>
            <Bezel.Inner className="px-5 py-4">
              <p className="text-xs text-white/30 uppercase tracking-widest mb-1">Side to move</p>
              <div className="flex items-center gap-2">
                <div
                  className={cn(
                    "w-4 h-4 rounded-full ring-1",
                    sideToMove === "white"
                      ? "bg-white ring-white/30"
                      : "bg-[#111] ring-white/20"
                  )}
                />
                <span className="text-sm font-medium capitalize">{sideToMove}</span>
              </div>
            </Bezel.Inner>
          </Bezel>

          {/* Move list */}
          <Bezel>
            <Bezel.Inner className="p-4">
              <p className="text-xs text-white/30 uppercase tracking-widest mb-3 px-2">Moves</p>
              <MoveList moves={wsState ? [] : moves} className="max-h-60" />
            </Bezel.Inner>
          </Bezel>

          {/* Actions */}
          {!isTerminal(status) && (
            <Bezel>
              <Bezel.Inner className="p-4 flex gap-3">
                <button
                  onClick={handleResign}
                  className="flex items-center gap-2 text-xs text-white/40 hover:text-rose-400/80 ring-1 ring-white/10 rounded-full px-4 py-2.5 hover:ring-rose-500/30 transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)]"
                >
                  <Flag size={14} weight="light" />
                  Resign
                </button>
              </Bezel.Inner>
            </Bezel>
          )}

          {/* Coach */}
          {isTerminal(status) && (
            <Bezel>
              <Bezel.Inner className="p-5">
                <button
                  onClick={() => setShowCoach((v) => !v)}
                  className="group w-full flex items-center gap-2 text-xs text-white/50 hover:text-white transition-colors duration-500 mb-4"
                >
                  <ChartLineUp size={14} weight="light" />
                  <span>{showCoach ? "Hide analysis" : "AI Coach analysis"}</span>
                </button>
                {showCoach && <CoachPanel gameId={gameId} />}
              </Bezel.Inner>
            </Bezel>
          )}

          {/* Invite code */}
          {game?.inviteCode && (
            <Bezel>
              <Bezel.Inner className="px-5 py-4">
                <p className="text-xs text-white/30 uppercase tracking-widest mb-1">Invite code</p>
                <p className="font-mono text-lg tracking-[0.25em] text-white/80">{game.inviteCode}</p>
              </Bezel.Inner>
            </Bezel>
          )}
        </Reveal>
      </div>
    </main>
  );
}
