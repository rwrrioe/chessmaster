import Link from "next/link";
import Reveal from "@/components/Reveal";
import Bezel from "@/components/Bezel";
import CTAButton from "@/components/CTAButton";
import EyebrowTag from "@/components/EyebrowTag";

export default function Home() {
  return (
    <main className="min-h-[100dvh] overflow-x-hidden">

      {/* ─── Hero ──────────────────────────────────────────────────────── */}
      <section className="min-h-[100dvh] flex flex-col items-center justify-center px-4 pt-28 pb-24 text-center">
        <Reveal delay={100}>
          <h1 className="font-display text-[clamp(3rem,10vw,8rem)] leading-[0.95] tracking-tight">
            Play Chess
            <br />
            <span className="font-editorial italic text-white/60">at its peak.</span>
          </h1>
        </Reveal>

        <Reveal delay={200}>
          <p className="mt-8 max-w-lg text-white/40 text-lg leading-relaxed">
            Real-time multiplayer, five AI difficulty levels, and Gemini-powered
            move-by-move coaching — all in a platform that feels like the future.
          </p>
        </Reveal>

        <Reveal delay={300}>
          <div className="mt-12 flex flex-wrap items-center justify-center gap-4">
            <CTAButton href="/register">Start playing free</CTAButton>
            <CTAButton href="/leaderboard" variant="ghost">
              See leaderboard
            </CTAButton>
          </div>
        </Reveal>

        {/* Mini board illustration */}
        <Reveal delay={400} className="mt-20 w-full max-w-sm mx-auto">
          <Bezel>
            <Bezel.Inner className="p-4">
              <div className="grid grid-cols-8 rounded-lg overflow-hidden ring-1 ring-white/5">
                {Array.from({ length: 64 }, (_, i) => {
                  const r = Math.floor(i / 8);
                  const f = i % 8;
                  const light = (r + f) % 2 !== 0;
                  const initials: Record<string, string> = {
                    "0-0": "♜","0-1": "♞","0-2": "♝","0-3": "♛","0-4": "♚","0-5": "♝","0-6": "♞","0-7": "♜",
                    "1-0": "♟","1-1": "♟","1-2": "♟","1-3": "♟","1-4": "♟","1-5": "♟","1-6": "♟","1-7": "♟",
                    "6-0": "♙","6-1": "♙","6-2": "♙","6-3": "♙","6-4": "♙","6-5": "♙","6-6": "♙","6-7": "♙",
                    "7-0": "♖","7-1": "♘","7-2": "♗","7-3": "♕","7-4": "♔","7-5": "♗","7-6": "♘","7-7": "♖",
                  };
                  const piece = initials[`${r}-${f}`];
                  return (
                    <div
                      key={i}
                      className={`aspect-square flex items-center justify-center text-xs ${
                        light ? "bg-[#1a1a1a]" : "bg-[#0a0a0a]"
                      }`}
                    >
                      {piece && (
                        <span
                          className={
                            r < 2
                              ? "text-white/30"
                              : "text-white/80"
                          }
                        >
                          {piece}
                        </span>
                      )}
                    </div>
                  );
                })}
              </div>
            </Bezel.Inner>
          </Bezel>
        </Reveal>
      </section>

      {/* ─── Asymmetric Bento ──────────────────────────────────────────── */}
      <section className="px-4 md:px-8 lg:px-16 py-24 max-w-7xl mx-auto">
        <Reveal>
          <EyebrowTag>Everything you need</EyebrowTag>
          <h2 className="mt-6 font-display text-4xl md:text-6xl tracking-tight">
            One platform.<br />
            <span className="text-white/40">Infinite depth.</span>
          </h2>
        </Reveal>

        {/* Bento grid */}
        <div className="mt-16 grid grid-cols-1 md:grid-cols-12 gap-4">

          {/* Live Multiplayer — large */}
          <Reveal className="md:col-span-7 md:row-span-2">
            <Bezel className="h-full">
              <Bezel.Inner className="p-8 md:p-10 h-full flex flex-col justify-between min-h-[320px]">
                <div>
                  <EyebrowTag>Multiplayer</EyebrowTag>
                  <h3 className="mt-4 font-display text-3xl md:text-4xl tracking-tight">
                    Live PvP via WebSockets
                  </h3>
                  <p className="mt-4 text-white/40 text-sm leading-relaxed max-w-sm">
                    Create a room, share the invite code. Your opponent joins and
                    moves propagate in real-time — sub-50ms latency.
                  </p>
                </div>
                <CTAButton href="/play" variant="ghost" className="self-start mt-8">
                  Create game
                </CTAButton>
              </Bezel.Inner>
            </Bezel>
          </Reveal>

          {/* AI Difficulty */}
          <Reveal delay={80} className="md:col-span-5">
            <Bezel className="h-full">
              <Bezel.Inner className="p-6 h-full min-h-[140px] flex flex-col justify-between">
                <EyebrowTag>AI Opponent</EyebrowTag>
                <div>
                  <h3 className="mt-3 font-display text-xl tracking-tight">Three difficulty tiers</h3>
                  <div className="mt-4 flex gap-2">
                    {["Easy", "Medium", "Hard"].map((lvl) => (
                      <span
                        key={lvl}
                        className="text-xs rounded-full px-3 py-1 ring-1 ring-white/10 bg-white/5 text-white/50"
                      >
                        {lvl}
                      </span>
                    ))}
                  </div>
                </div>
              </Bezel.Inner>
            </Bezel>
          </Reveal>

          {/* Gemini Coach */}
          <Reveal delay={120} className="md:col-span-5">
            <Bezel className="h-full">
              <Bezel.Inner className="p-6 h-full min-h-[140px] flex flex-col justify-between">
                <EyebrowTag>AI Coach</EyebrowTag>
                <div>
                  <h3 className="mt-3 font-display text-xl tracking-tight">Gemini post-game analysis</h3>
                  <p className="mt-2 text-white/35 text-xs leading-relaxed">
                    Blunders, mistakes, inaccuracies — ranked by severity with better alternatives.
                  </p>
                </div>
              </Bezel.Inner>
            </Bezel>
          </Reveal>

          {/* Leaderboard */}
          <Reveal delay={60} className="md:col-span-4">
            <Bezel className="h-full">
              <Bezel.Inner className="p-6 h-full min-h-[180px] flex flex-col justify-between">
                <EyebrowTag>Rankings</EyebrowTag>
                <div>
                  <h3 className="mt-3 font-display text-2xl tracking-tight">City leaderboard</h3>
                  <p className="mt-2 text-white/35 text-xs leading-relaxed">
                    Elo-based global rankings, filterable by city.
                  </p>
                </div>
                <CTAButton href="/leaderboard" variant="ghost" className="self-start mt-6">
                  View rankings
                </CTAButton>
              </Bezel.Inner>
            </Bezel>
          </Reveal>

          {/* Pro badge */}
          <Reveal delay={160} className="md:col-span-8">
            <Bezel
              className="h-full"
              style={{
                background:
                  "radial-gradient(ellipse at 80% 50%, rgba(91,33,182,0.12) 0%, transparent 70%)",
              } as React.CSSProperties}
            >
              <Bezel.Inner className="p-8 h-full min-h-[180px] flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
                <div>
                  <EyebrowTag className="ring-purple-500/30 text-purple-300 bg-purple-500/10">
                    Pro membership
                  </EyebrowTag>
                  <h3 className="mt-4 font-display text-2xl md:text-3xl tracking-tight">
                    Unlock unlimited coaching
                  </h3>
                  <p className="mt-2 text-white/35 text-sm">
                    Priority analysis, extended history, and more coming soon.
                  </p>
                </div>
                <CTAButton href="/me">Upgrade to Pro</CTAButton>
              </Bezel.Inner>
            </Bezel>
          </Reveal>
        </div>
      </section>

      {/* ─── Footer ────────────────────────────────────────────────────── */}
      <footer className="py-16 px-4 text-center border-t border-white/5">
        <p className="text-white/20 text-xs tracking-widest uppercase">
          ChessMaster Pro &mdash; Phase 5
        </p>
      </footer>
    </main>
  );
}
