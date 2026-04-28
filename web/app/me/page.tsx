"use client";
import { useEffect, useState } from "react";
import { me, myGames, upgradePro, getToken } from "@/lib/api";
import type { Player, Game } from "@/lib/types";
import Bezel from "@/components/Bezel";
import EyebrowTag from "@/components/EyebrowTag";
import CTAButton from "@/components/CTAButton";
import Reveal from "@/components/Reveal";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/cn";

function GameRow({ game }: { game: Game }) {
  const statusColor = game.status === "active"
    ? "text-emerald-400/70"
    : game.status.includes("won") || game.status === "draw"
    ? "text-white/40"
    : "text-white/30";

  return (
    <a href={`/play/${game.id}`} className="block">
      <Bezel className="hover:ring-white/20 transition-all duration-500">
        <Bezel.Inner className="px-5 py-4 flex items-center gap-4">
          <span className="text-xs font-mono text-white/20 uppercase tracking-widest w-16">
            {game.mode}
          </span>
          <span className={cn("text-xs flex-1", statusColor)}>{game.status}</span>
          {game.result && (
            <span className="font-mono text-sm text-white/50">{game.result}</span>
          )}
        </Bezel.Inner>
      </Bezel>
    </a>
  );
}

export default function MePage() {
  const router = useRouter();
  const [player, setPlayer] = useState<Player | null>(null);
  const [games, setGames] = useState<Game[]>([]);
  const [loading, setLoading] = useState(true);
  const [upgrading, setUpgrading] = useState(false);
  const [upgradeError, setUpgradeError] = useState<string | null>(null);

  useEffect(() => {
    if (!getToken()) {
      router.push("/login");
      return;
    }
    Promise.all([me(), myGames()])
      .then(([p, g]) => {
        setPlayer(p);
        setGames(g ?? []);
      })
      .catch(() => router.push("/login"))
      .finally(() => setLoading(false));
  }, [router]);

  async function handleUpgrade() {
    setUpgrading(true);
    setUpgradeError(null);
    try {
      const updated = await upgradePro();
      setPlayer(updated);
    } catch (err) {
      setUpgradeError(err instanceof Error ? err.message : "Upgrade failed");
    } finally {
      setUpgrading(false);
    }
  }

  if (loading) {
    return (
      <main className="min-h-[100dvh] flex items-center justify-center">
        <span className="w-8 h-8 rounded-full border border-white/20 border-t-white/70 animate-spin" />
      </main>
    );
  }

  if (!player) return null;

  return (
    <main className="min-h-[100dvh] px-4 pt-32 pb-24 max-w-3xl mx-auto">
      <Reveal>
        <div className="flex items-center gap-4">
          {/* Avatar */}
          <div className="w-16 h-16 rounded-full bg-white/5 ring-1 ring-white/10 flex items-center justify-center">
            <span className="font-display text-2xl text-white/70">
              {player.username[0].toUpperCase()}
            </span>
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="font-display text-3xl tracking-tight">{player.username}</h1>
              {player.isPro && (
                <EyebrowTag className="ring-purple-500/30 text-purple-300 bg-purple-500/10">
                  Pro
                </EyebrowTag>
              )}
            </div>
            <p className="text-white/40 text-sm mt-0.5">{player.email}</p>
          </div>
        </div>
      </Reveal>

      {/* Stats */}
      <Reveal delay={80} className="mt-10 grid grid-cols-2 md:grid-cols-3 gap-3">
        {[
          { label: "City", value: player.city || "—" },
          { label: "Games played", value: String(games.length) },
          { label: "Member since", value: new Date(player.createdAt).getFullYear().toString() },
        ].map(({ label, value }) => (
          <Bezel key={label}>
            <Bezel.Inner className="p-5">
              <p className="text-xs text-white/30 uppercase tracking-widest mb-2">{label}</p>
              <p className="font-display text-xl text-white/90">{value}</p>
            </Bezel.Inner>
          </Bezel>
        ))}
      </Reveal>

      {/* Pro upgrade */}
      {!player.isPro && (
        <Reveal delay={120} className="mt-8">
          <Bezel
            style={{
              background:
                "radial-gradient(ellipse at 80% 50%, rgba(91,33,182,0.10) 0%, transparent 70%)",
            } as React.CSSProperties}
          >
            <Bezel.Inner className="p-8 flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
              <div>
                <EyebrowTag className="ring-purple-500/30 text-purple-300 bg-purple-500/10">
                  Pro plan
                </EyebrowTag>
                <h2 className="mt-3 font-display text-2xl tracking-tight">Unlock unlimited coaching</h2>
                <p className="mt-2 text-white/40 text-sm leading-relaxed">
                  Priority Gemini analysis, extended history, and more.
                </p>
                {upgradeError && (
                  <p className="mt-3 text-xs text-rose-400/70">{upgradeError}</p>
                )}
              </div>
              <CTAButton onClick={handleUpgrade} disabled={upgrading}>
                {upgrading ? "Upgrading…" : "Upgrade to Pro"}
              </CTAButton>
            </Bezel.Inner>
          </Bezel>
        </Reveal>
      )}

      {/* Recent games */}
      <Reveal delay={160} className="mt-10">
        <p className="text-xs text-white/30 uppercase tracking-widest mb-4">Recent games</p>
        {games.length === 0 && (
          <p className="text-white/20 text-sm text-center py-8">No games yet.</p>
        )}
        <div className="space-y-2">
          {games.slice(0, 10).map((g) => (
            <GameRow key={g.id} game={g} />
          ))}
        </div>
      </Reveal>
    </main>
  );
}
