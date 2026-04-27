export default function Home() {
  return (
    <main className="min-h-[100dvh] flex items-center justify-center px-4 py-24">
      <section className="max-w-3xl text-center">
        <span className="inline-block rounded-full px-3 py-1 text-[10px] uppercase tracking-[0.2em] font-medium ring-1 ring-white/10 bg-white/5">
          Phase 1 — Foundation
        </span>
        <h1 className="mt-8 font-display text-6xl md:text-8xl tracking-tight">
          ChessMaster <span className="italic font-editorial">Pro</span>
        </h1>
        <p className="mt-6 text-white/60 text-lg">
          Premium UI, real-time multiplayer, and Gemini coaching arrive in later phases.
        </p>
      </section>
    </main>
  );
}
