"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { createGame, joinGame } from "@/lib/api";
import Bezel from "@/components/Bezel";
import EyebrowTag from "@/components/EyebrowTag";
import CTAButton from "@/components/CTAButton";
import Reveal from "@/components/Reveal";
import { cn } from "@/lib/cn";

type Mode = "pvp_create" | "pvp_join" | "ai_easy" | "ai_medium" | "ai_hard";

const MODES: Array<{ id: Mode; label: string; description: string; eyebrow: string }> = [
  { id: "pvp_create", label: "Create PvP", description: "Generate an invite link to play against a friend.", eyebrow: "Multiplayer" },
  { id: "pvp_join", label: "Join by code", description: "Enter a friend's invite code to jump into their game.", eyebrow: "Multiplayer" },
  { id: "ai_easy", label: "AI — Easy", description: "Perfect for beginners. The engine plays casually.", eyebrow: "Solo" },
  { id: "ai_medium", label: "AI — Medium", description: "A balanced challenge for intermediate players.", eyebrow: "Solo" },
  { id: "ai_hard", label: "AI — Hard", description: "Engine plays near-optimal moves. Come prepared.", eyebrow: "Solo" },
];

export default function PlayPage() {
  const router = useRouter();
  const [selected, setSelected] = useState<Mode>("pvp_create");
  const [color, setColor] = useState<"white" | "black" | "random">("white");
  const [inviteCode, setInviteCode] = useState("");
  const [generatedCode, setGeneratedCode] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleStart() {
    setError(null);
    setLoading(true);
    try {
      if (selected === "pvp_join") {
        const game = await joinGame(inviteCode.trim());
        router.push(`/play/${game.id}`);
        return;
      }

      const mode = selected === "pvp_create" ? "pvp" : selected;
      const game = await createGame(mode, color);

      if (mode === "pvp" && game.inviteCode) {
        setGeneratedCode(game.inviteCode);
      }
      router.push(`/play/${game.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start game");
    } finally {
      setLoading(false);
    }
  }

  const modeInfo = MODES.find((m) => m.id === selected)!;

  return (
    <main className="min-h-[100dvh] px-4 pt-32 pb-24 max-w-4xl mx-auto">
      <Reveal>
        <EyebrowTag>New game</EyebrowTag>
        <h1 className="mt-6 font-display text-5xl md:text-7xl tracking-tight">
          Choose your battle.
        </h1>
      </Reveal>

      <div className="mt-16 grid grid-cols-1 md:grid-cols-12 gap-4">
        {/* Mode selector */}
        <Reveal delay={80} className="md:col-span-5">
          <Bezel className="h-full">
            <Bezel.Inner className="p-3 space-y-1.5 h-full">
              {MODES.map((m) => (
                <button
                  key={m.id}
                  onClick={() => setSelected(m.id)}
                  className={cn(
                    "w-full text-left px-4 py-3 rounded-[calc(2rem-0.375rem-0.75rem)] text-sm transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)]",
                    selected === m.id
                      ? "bg-white/10 text-white ring-1 ring-white/20"
                      : "text-white/40 hover:text-white/70 hover:bg-white/5"
                  )}
                >
                  <span className="text-[10px] uppercase tracking-[0.15em] text-white/30 block mb-0.5">
                    {m.eyebrow}
                  </span>
                  {m.label}
                </button>
              ))}
            </Bezel.Inner>
          </Bezel>
        </Reveal>

        {/* Configuration panel */}
        <Reveal delay={140} className="md:col-span-7">
          <Bezel className="h-full">
            <Bezel.Inner className="p-8 h-full flex flex-col justify-between min-h-[340px]">
              <div>
                <EyebrowTag>{modeInfo.eyebrow}</EyebrowTag>
                <h2 className="mt-4 font-display text-2xl tracking-tight">{modeInfo.label}</h2>
                <p className="mt-3 text-white/40 text-sm leading-relaxed">{modeInfo.description}</p>

                {/* Color picker — only for non-join modes */}
                {selected !== "pvp_join" && (
                  <div className="mt-8 space-y-3">
                    <p className="text-xs text-white/30 uppercase tracking-widest">Play as</p>
                    <div className="flex gap-2">
                      {(["white", "black", "random"] as const).map((c) => (
                        <button
                          key={c}
                          onClick={() => setColor(c)}
                          className={cn(
                            "flex-1 py-2.5 rounded-full text-xs capitalize transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)]",
                            color === c
                              ? "bg-white text-black"
                              : "ring-1 ring-white/10 bg-white/5 text-white/50 hover:bg-white/10"
                          )}
                        >
                          {c}
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                {/* Invite code input — join mode */}
                {selected === "pvp_join" && (
                  <div className="mt-8 space-y-2">
                    <label className="text-xs text-white/30 uppercase tracking-widest">
                      Invite Code
                    </label>
                    <input
                      value={inviteCode}
                      onChange={(e) => setInviteCode(e.target.value)}
                      placeholder="ABCD-1234"
                      className="w-full rounded-2xl bg-white/5 ring-1 ring-white/10 px-4 py-3 text-sm text-white placeholder-white/20 outline-none focus:ring-white/30 font-mono tracking-widest transition-all duration-500"
                    />
                  </div>
                )}

                {generatedCode && (
                  <div className="mt-6 p-4 rounded-2xl ring-1 ring-emerald-500/20 bg-emerald-500/5">
                    <p className="text-xs text-emerald-400/60 uppercase tracking-widest mb-1">Invite code</p>
                    <p className="font-mono text-xl text-emerald-400 tracking-[0.3em]">{generatedCode}</p>
                  </div>
                )}

                {error && (
                  <p className="mt-4 text-xs text-rose-400/80 py-2 ring-1 ring-rose-500/20 rounded-xl bg-rose-500/5 px-3 text-center">
                    {error}
                  </p>
                )}
              </div>

              <CTAButton
                onClick={handleStart}
                disabled={loading || (selected === "pvp_join" && !inviteCode.trim())}
                className="mt-8 self-start"
              >
                {loading ? "Starting…" : selected === "pvp_join" ? "Join game" : "Start game"}
              </CTAButton>
            </Bezel.Inner>
          </Bezel>
        </Reveal>
      </div>
    </main>
  );
}
