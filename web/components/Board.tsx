"use client";
import { useState, useCallback } from "react";
import Bezel from "./Bezel";
import { cn } from "@/lib/cn";

// Unicode chess glyphs: indexed by piece letter (uppercase = white)
const PIECE_GLYPHS: Record<string, string> = {
  K: "♔", Q: "♕", R: "♖", B: "♗", N: "♘", P: "♙",
  k: "♚", q: "♛", r: "♜", b: "♝", n: "♞", p: "♟",
};

const FILES = ["a", "b", "c", "d", "e", "f", "g", "h"];
const RANKS = [8, 7, 6, 5, 4, 3, 2, 1];

/**
 * Parse FEN position string into a 8×8 board array (rank 8 = index 0).
 * Returns array[rank][file] = piece char or null.
 */
function parseFEN(fen: string): (string | null)[][] {
  const board: (string | null)[][] = Array.from({ length: 8 }, () =>
    Array(8).fill(null)
  );
  const pos = fen.split(" ")[0];
  const rows = pos.split("/");
  for (let r = 0; r < 8; r++) {
    let f = 0;
    for (const ch of rows[r]) {
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

/**
 * 8×8 chess board rendered with unicode pieces (premium, zero asset weight).
 * Light squares #1a1a1a, dark squares #0a0a0a.
 * Drag source: mousedown sets dragging piece; drop target: mouseup validates UCI.
 * Wrapped in Double-Bezel.
 */
export default function Board({
  fen,
  legalMoves,
  onMove,
  orientation = "white",
  disabled = false,
}: BoardProps) {
  const board = parseFEN(fen);
  const [dragging, setDragging] = useState<{ r: number; f: number } | null>(null);
  const [hover, setHover] = useState<{ r: number; f: number } | null>(null);

  const displayRanks = orientation === "white" ? RANKS : [...RANKS].reverse();
  const displayFiles = orientation === "white" ? FILES : [...FILES].reverse();

  const legalSet = new Set(legalMoves);

  const isLegalTarget = useCallback(
    (fromR: number, fromF: number, toR: number, toF: number) => {
      const uci = toUCI(fromF, fromR, toF, toR);
      return legalSet.has(uci) || legalSet.has(uci + "q"); // promotion
    },
    [legalSet]
  );

  function handleDragStart(r: number, f: number) {
    if (disabled) return;
    const piece = board[r][f];
    if (!piece) return;
    setDragging({ r, f });
  }

  function handleDrop(r: number, f: number) {
    if (!dragging) return;
    const { r: fr, f: ff } = dragging;
    if (fr === r && ff === f) {
      setDragging(null);
      return;
    }
    const uci = toUCI(ff, fr, f, r);
    const uciQ = uci + "q";
    if (legalSet.has(uci)) {
      onMove(uci);
    } else if (legalSet.has(uciQ)) {
      onMove(uciQ);
    }
    setDragging(null);
    setHover(null);
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
        <div
          className="grid grid-cols-8 ring-1 ring-white/5 rounded-lg overflow-hidden select-none"
          onMouseLeave={() => {
            setDragging(null);
            setHover(null);
          }}
        >
          {displayRanks.map((rank) => {
            const ri = rankIndex(rank);
            return displayFiles.map((file) => {
              const fi = fileIndex(file);
              const piece = board[ri][fi];
              const isLight = (ri + fi) % 2 !== 0;
              const isDragged = dragging?.r === ri && dragging?.f === fi;
              const isHovered = hover?.r === ri && hover?.f === fi;
              const canDrop =
                dragging && isLegalTarget(dragging.r, dragging.f, ri, fi);

              return (
                <div
                  key={`${ri}-${fi}`}
                  className={cn(
                    "aspect-square flex items-center justify-center relative cursor-pointer",
                    "transition-colors duration-200",
                    isLight ? "bg-[#1a1a1a]" : "bg-[#0a0a0a]",
                    isDragged && "opacity-40",
                    canDrop && isHovered && "bg-white/20",
                    canDrop && !isHovered && "after:absolute after:inset-[30%] after:rounded-full after:bg-white/20"
                  )}
                  onMouseDown={() => handleDragStart(ri, fi)}
                  onMouseEnter={() => {
                    if (dragging) setHover({ r: ri, f: fi });
                  }}
                  onMouseUp={() => handleDrop(ri, fi)}
                >
                  {piece && (
                    <span
                      className={cn(
                        "text-2xl leading-none select-none transition-all duration-300 ease-[cubic-bezier(0.32,0.72,0,1)]",
                        piece === piece.toUpperCase()
                          ? "text-white drop-shadow-[0_1px_2px_rgba(0,0,0,0.8)]"
                          : "text-white/30 drop-shadow-[0_1px_2px_rgba(0,0,0,0.8)]",
                        dragging?.r === ri && dragging?.f === fi && "scale-110"
                      )}
                    >
                      {PIECE_GLYPHS[piece] ?? piece}
                    </span>
                  )}
                  {/* Rank label on leftmost file */}
                  {fi === (orientation === "white" ? 0 : 7) && (
                    <span className="absolute top-0.5 left-0.5 text-[8px] text-white/20 font-mono leading-none">
                      {rank}
                    </span>
                  )}
                  {/* File label on bottom rank */}
                  {ri === (orientation === "white" ? 7 : 0) && (
                    <span className="absolute bottom-0.5 right-0.5 text-[8px] text-white/20 font-mono leading-none">
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
