"use client";
import { useState, useEffect } from "react";
import { leaderboard } from "@/lib/api";
import type { LeaderboardEntry } from "@/lib/types";
import LeaderboardCard from "@/components/LeaderboardCard";
import EyebrowTag from "@/components/EyebrowTag";
import Reveal from "@/components/Reveal";
import { MagnifyingGlass } from "phosphor-react";

export default function LeaderboardPage() {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [city, setCity] = useState("");
  const [loading, setLoading] = useState(true);
  const [inputCity, setInputCity] = useState("");

  useEffect(() => {
    setLoading(true);
    leaderboard(city || undefined, 50)
      .then((d) => setEntries(d ?? []))
      .catch(() => setEntries([]))
      .finally(() => setLoading(false));
  }, [city]);

  function applyFilter() {
    setCity(inputCity.trim());
  }

  return (
    <main className="min-h-[100dvh] px-4 pt-32 pb-24 max-w-4xl mx-auto">
      <Reveal>
        <EyebrowTag>Rankings</EyebrowTag>
        <h1 className="mt-6 font-display text-5xl md:text-7xl tracking-tight">
          Leaderboard.
        </h1>
        <p className="mt-4 text-white/40 text-lg">
          Top players by Elo rating, optionally filtered by city.
        </p>
      </Reveal>

      {/* City filter */}
      <Reveal delay={80} className="mt-12">
        <div className="flex gap-3">
          <div className="flex-1 flex items-center gap-3 rounded-full bg-white/5 ring-1 ring-white/10 px-5 py-3">
            <MagnifyingGlass size={14} weight="light" className="text-white/30 shrink-0" />
            <input
              value={inputCity}
              onChange={(e) => setInputCity(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && applyFilter()}
              placeholder="Filter by city…"
              className="flex-1 bg-transparent text-sm text-white placeholder-white/20 outline-none"
            />
          </div>
          <button
            onClick={applyFilter}
            className="rounded-full px-5 py-3 text-sm bg-white text-black hover:bg-white/90 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]"
          >
            Filter
          </button>
          {city && (
            <button
              onClick={() => { setCity(""); setInputCity(""); }}
              className="rounded-full px-4 py-3 text-sm ring-1 ring-white/10 bg-white/5 text-white/50 hover:text-white transition-all duration-500"
            >
              Clear
            </button>
          )}
        </div>
      </Reveal>

      {/* Grid */}
      <div className="mt-10 space-y-3">
        {loading && (
          <div className="flex items-center justify-center py-16">
            <span className="w-6 h-6 rounded-full border border-white/20 border-t-white/60 animate-spin" />
          </div>
        )}

        {!loading && entries.length === 0 && (
          <p className="text-center text-white/30 py-16">No players found.</p>
        )}

        {!loading &&
          entries.map((entry, i) => (
            <Reveal key={`${entry.username}-${i}`} delay={i * 30}>
              <LeaderboardCard entry={entry} rank={i + 1} />
            </Reveal>
          ))}
      </div>
    </main>
  );
}
