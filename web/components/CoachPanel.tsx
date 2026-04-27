"use client";
import { useState } from "react";
import { Brain } from "phosphor-react";
import { coach } from "@/lib/api";
import type { Analysis, Mistake } from "@/lib/types";
import Bezel from "./Bezel";
import { cn } from "@/lib/cn";

interface CoachPanelProps {
  gameId: string;
}

const severityStyles: Record<Mistake["severity"], string> = {
  inaccuracy: "bg-cyan-500/20 text-cyan-300 ring-cyan-500/30",
  mistake: "bg-amber-500/30 text-amber-300 ring-amber-500/40",
  blunder: "bg-rose-500/40 text-rose-300 ring-rose-500/50",
};

/**
 * Calls /games/{id}/coach on demand.
 * Displays summary + per-mistake cards with severity chips.
 * Each mistake uses Double-Bezel.
 */
export default function CoachPanel({ gameId }: CoachPanelProps) {
  const [analysis, setAnalysis] = useState<Analysis | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function requestCoach() {
    setLoading(true);
    setError(null);
    try {
      const data = await coach(gameId);
      setAnalysis(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Coach unavailable");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-4">
      {!analysis && !loading && (
        <button
          onClick={requestCoach}
          className="group w-full flex items-center justify-center gap-3 rounded-full py-3 px-5 ring-1 ring-white/10 bg-white/5 text-sm text-white/70 hover:text-white hover:bg-white/10 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]"
        >
          <Brain size={16} weight="light" />
          Analyse with Gemini Coach
        </button>
      )}

      {loading && (
        <div className="flex items-center justify-center gap-3 py-6 text-white/40 text-sm">
          <span className="w-4 h-4 rounded-full border border-white/20 border-t-white/70 animate-spin" />
          Analysing game…
        </div>
      )}

      {error && (
        <p className="text-rose-400/70 text-xs text-center py-3 ring-1 ring-rose-500/20 rounded-xl bg-rose-500/5 px-4">
          {error}
        </p>
      )}

      {analysis && (
        <div className="space-y-4">
          <p className="text-sm text-white/70 leading-relaxed">{analysis.summary}</p>

          {analysis.mistakes.length === 0 && (
            <p className="text-xs text-white/40 text-center py-3">No significant mistakes found.</p>
          )}

          {analysis.mistakes.map((m, i) => (
            <Bezel key={i}>
              <Bezel.Inner className="p-4 space-y-2">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-xs text-white/30">
                    Move {m.ply}
                  </span>
                  <span
                    className={cn(
                      "rounded-full px-2 py-0.5 text-[10px] uppercase tracking-[0.15em] font-medium ring-1",
                      severityStyles[m.severity]
                    )}
                  >
                    {m.severity}
                  </span>
                </div>
                <div className="flex items-center gap-3 text-sm">
                  <span className="font-mono text-white/50 line-through">{m.move}</span>
                  <span className="text-white/30 text-xs">→</span>
                  <span className="font-mono text-emerald-400/80">{m.better}</span>
                </div>
                <p className="text-xs text-white/50 leading-relaxed">{m.comment}</p>
              </Bezel.Inner>
            </Bezel>
          ))}

          <button
            onClick={() => setAnalysis(null)}
            className="text-xs text-white/30 hover:text-white/60 transition-colors duration-300 mx-auto block"
          >
            Reset
          </button>
        </div>
      )}
    </div>
  );
}
