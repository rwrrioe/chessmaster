import type { Move } from "@/lib/types";
import { cn } from "@/lib/cn";

interface MoveListProps {
  moves: Move[];
  className?: string;
}

/**
 * Vertical SAN move pair list with monospace numerals.
 * Pairs: white move + black move per row, styled subtly.
 */
export default function MoveList({ moves, className }: MoveListProps) {
  const pairs: Array<{ num: number; white: Move; black?: Move }> = [];
  for (let i = 0; i < moves.length; i += 2) {
    pairs.push({
      num: Math.floor(i / 2) + 1,
      white: moves[i],
      black: moves[i + 1],
    });
  }

  return (
    <div className={cn("overflow-y-auto", className)}>
      {pairs.length === 0 && (
        <p className="text-white/30 text-xs py-4 text-center">No moves yet</p>
      )}
      {pairs.map(({ num, white, black }) => (
        <div
          key={num}
          className="grid grid-cols-[2rem_1fr_1fr] gap-x-2 px-3 py-1.5 rounded-lg hover:bg-white/5 transition-colors duration-300 group"
        >
          <span className="font-mono text-xs text-white/25 self-center tabular-nums">
            {num}.
          </span>
          <span className="font-mono text-xs text-white/80 hover:text-white transition-colors duration-300 cursor-default">
            {white.san}
          </span>
          {black && (
            <span className="font-mono text-xs text-white/50 hover:text-white/80 transition-colors duration-300 cursor-default">
              {black.san}
            </span>
          )}
        </div>
      ))}
    </div>
  );
}
