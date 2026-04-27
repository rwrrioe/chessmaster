import type { LeaderboardEntry } from "@/lib/types";
import Bezel from "./Bezel";
import { cn } from "@/lib/cn";

interface LeaderboardCardProps {
  entry: LeaderboardEntry;
  rank: number;
  className?: string;
}

const rankStyles: Record<number, string> = {
  1: "text-yellow-400",
  2: "text-white/60",
  3: "text-amber-600",
};

/**
 * Bento entry card for leaderboard. Double-Bezel wrapped.
 * Rank 1–3 get special colour treatment.
 */
export default function LeaderboardCard({
  entry,
  rank,
  className,
}: LeaderboardCardProps) {
  return (
    <Bezel className={cn(className)}>
      <Bezel.Inner className="px-5 py-4 flex items-center gap-4">
        {/* Rank */}
        <span
          className={cn(
            "font-mono text-2xl font-semibold w-10 shrink-0 tabular-nums",
            rankStyles[rank] ?? "text-white/20"
          )}
        >
          {rank}
        </span>

        {/* Avatar placeholder */}
        <div className="w-9 h-9 rounded-full bg-white/5 ring-1 ring-white/10 flex items-center justify-center shrink-0">
          <span className="text-sm font-display text-white/60">
            {entry.username[0].toUpperCase()}
          </span>
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-white/90 truncate">{entry.username}</p>
          <p className="text-[11px] text-white/30 truncate">{entry.city}</p>
        </div>

        {/* Stats */}
        <div className="text-right shrink-0">
          <p className="text-base font-display text-white/90 tabular-nums">{entry.elo}</p>
          <p className="text-[10px] text-white/30 tabular-nums">
            {entry.wins}W / {entry.games - entry.wins}
          </p>
        </div>
      </Bezel.Inner>
    </Bezel>
  );
}
