"use client";
import { useState, useCallback } from "react";
import Bezel from "./Bezel";
import { cn } from "@/lib/cn";

const PIECE_GLYPHS: Record<string, string> = {
  K: "♔", Q: "♕", R: "♖", B: "♗", N: "♘", P: "♙",
  k: "♚", q: "♛", r: "♜", b: "♝", n: "♞", p: "♟",
};

const FILES = ["a", "b", "c", "d", "e", "f", "g", "h"];
const RANKS = [8, 7, 6, 5, 4, 3, 2, 1];

function parseFEN(fen: string): (string | null)[][] {
  const board: (string | null)[][] = Array.from({ length: 8 }, () =>
    Array(8).fill(null)
  );
  const pos = fen.split(" ")[0];
  const rows = pos.split("/");
  for (let r = 0; r < 8; r++) {
    let f = 0;
    for (const ch of rows[r] ?? "") {
      if (/\d/.test(ch)) {
        f += parseInt(ch, 10);
      } else {
        board[r][f] = ch;
        f++;
      }
    }
  }
  return board;
}

function toUCI(fromFile: number, fromRank: number, toFile: number, toRank: number): string {
  return `${FILES[fromFile]}${8 - fromRank}${FILES[toFile]}${8 - toRank}`;
}

interface BoardProps {
  fen: string;
  legalMoves: string[];
  onMove: (uci: string) => void;
  orientation?: "white" | "black";
  disabled?: boolean;
}

export default function Board({
  fen,
  legalMoves,
  onMove,
  orientation = "white",
  disabled = false,
}: BoardProps) {
  const board = parseFEN(fen);
  const [selected, setSelected] = useState<{ r: number; f: number } | null>(null);

  const displayRanks = orientation === "white" ? RANKS : [...RANKS].reverse();
  const displayFiles = orientation === "white" ? FILES : [...FILES].reverse();

  const legalSet = new Set(legalMoves);

  const isLegalTarget = useCallback(
    (fromR: number, fromF: number, toR: number, toF: number) => {
      const uci = toUCI(fromF, fromR, toF, toR);
      return legalSet.has(uci) || legalSet.has(uci + "q");
    },
    [legalSet]
  );

  function attemptMove(fromR: number, fromF: number, toR: number, toF: number) {
    const uci = toUCI(fromF, fromR, toF, toR);
    if (legalSet.has(uci)) {
      onMove(uci);
      return true;
    }
    if (legalSet.has(uci + "q")) {
      onMove(uci + "q");
      return true;
    }
    return false;
  }

  function handleSquareClick(r: number, f: number) {
    if (disabled) return;
    const piece = board[r][f];

    if (selected) {
      // Same square — deselect
      if (selected.r === r && selected.f === f) {
        setSelected(null);
        return;
      }
      // Try the move
      if (attemptMove(selected.r, selected.f, r, f)) {
        setSelected(null);
        return;
      }
      // Clicked another own piece — switch selection
      if (piece) {
        setSelected({ r, f });
        return;
      }
      // Empty square, illegal — clear
      setSelected(null);
      return;
    }

    if (!piece) return;
    // Only allow selecting a piece that has at least one legal move from here
    const fromUCI = `${FILES[f]}${8 - r}`;
    const hasLegal = legalMoves.some((m) => m.startsWith(fromUCI));
    if (!hasLegal) return;
    setSelected({ r, f });
  }

  function handleDragStart(e: React.DragEvent, r: number, f: number) {
    if (disabled || !board[r][f]) {
      e.preventDefault();
      return;
    }
    setSelected({ r, f });
    // Required for Firefox drag to actually start
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", `${r},${f}`);
  }

  function handleDrop(e: React.DragEvent, r: number, f: number) {
    e.preventDefault();
    const data = e.dataTransfer.getData("text/plain");
    if (!data) return;
    const [fr, ff] = data.split(",").map((n) => parseInt(n, 10));
    if (Number.isNaN(fr) || Number.isNaN(ff)) return;
    if (attemptMove(fr, ff, r, f)) {
      setSelected(null);
    }
  }

  function handleDragOver(e: React.DragEvent) {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
  }

  function rankIndex(displayRank: number) {
    return 8 - displayRank;
  }
  function fileIndex(displayFile: string) {
    return FILES.indexOf(displayFile);
  }

  return (
    <Bezel>
      <Bezel.Inner className="p-3">
        <div className="grid grid-cols-8 ring-1 ring-white/5 rounded-lg overflow-hidden select-none">
          {displayRanks.map((rank) => {
            const ri = rankIndex(rank);
            return displayFiles.map((file) => {
              const fi = fileIndex(file);
              const piece = board[ri][fi];
              const isLight = (ri + fi) % 2 !== 0;
              const isSelected = selected?.r === ri && selected?.f === fi;
              const canDrop = selected && isLegalTarget(selected.r, selected.f, ri, fi);

              return (
                <div
                  key={`${ri}-${fi}`}
                  className={cn(
                    "aspect-square flex items-center justify-center relative cursor-pointer",
                    "transition-colors duration-200",
                    isLight ? "bg-[#1a1a1a]" : "bg-[#0a0a0a]",
                    isSelected && "ring-2 ring-amber-400/60 ring-inset",
                    canDrop && "bg-emerald-500/10",
                    canDrop && piece && "ring-2 ring-rose-400/40 ring-inset"
                  )}
                  onClick={() => handleSquareClick(ri, fi)}
                  onDragOver={handleDragOver}
                  onDrop={(e) => handleDrop(e, ri, fi)}
                >
                  {piece && (
                    <span
                      draggable={!disabled}
                      onDragStart={(e) => handleDragStart(e, ri, fi)}
                      className={cn(
                        "text-3xl leading-none select-none transition-transform duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] cursor-grab active:cursor-grabbing",
                        piece === piece.toUpperCase()
                          ? "text-white drop-shadow-[0_1px_2px_rgba(0,0,0,0.8)]"
                          : "text-white/40 drop-shadow-[0_1px_2px_rgba(0,0,0,0.8)]",
                        isSelected && "scale-110"
                      )}
                    >
                      {PIECE_GLYPHS[piece] ?? piece}
                    </span>
                  )}
                  {/* Legal-move dot when selected */}
                  {canDrop && !piece && (
                    <span className="absolute w-2.5 h-2.5 rounded-full bg-white/30 pointer-events-none" />
                  )}
                  {fi === (orientation === "white" ? 0 : 7) && (
                    <span className="absolute top-0.5 left-1 text-[8px] text-white/20 font-mono leading-none pointer-events-none">
                      {rank}
                    </span>
                  )}
                  {ri === (orientation === "white" ? 7 : 0) && (
                    <span className="absolute bottom-0.5 right-1 text-[8px] text-white/20 font-mono leading-none pointer-events-none">
                      {file}
                    </span>
                  )}
                </div>
              );
            });
          })}
        </div>
      </Bezel.Inner>
    </Bezel>
  );
}
