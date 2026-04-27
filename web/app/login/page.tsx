"use client";
import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { login, setToken } from "@/lib/api";
import Bezel from "@/components/Bezel";
import EyebrowTag from "@/components/EyebrowTag";
import CTAButton from "@/components/CTAButton";
import Reveal from "@/components/Reveal";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const { token } = await login(email, password);
      setToken(token);
      router.push("/play");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-[100dvh] flex items-center justify-center px-4 py-28">
      <Reveal className="w-full max-w-sm">
        <Bezel>
          <Bezel.Inner className="p-8 space-y-6">
            <div className="text-center space-y-3">
              <EyebrowTag>Welcome back</EyebrowTag>
              <h1 className="font-display text-3xl tracking-tight">Sign in</h1>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-1.5">
                <label className="text-xs text-white/40 tracking-wide uppercase">Email</label>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full rounded-2xl bg-white/5 ring-1 ring-white/10 px-4 py-3 text-sm text-white placeholder-white/20 outline-none focus:ring-white/30 transition-all duration-500"
                  placeholder="you@example.com"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs text-white/40 tracking-wide uppercase">Password</label>
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full rounded-2xl bg-white/5 ring-1 ring-white/10 px-4 py-3 text-sm text-white placeholder-white/20 outline-none focus:ring-white/30 transition-all duration-500"
                  placeholder="••••••••"
                />
              </div>

              {error && (
                <p className="text-xs text-rose-400/80 text-center py-2 ring-1 ring-rose-500/20 rounded-xl bg-rose-500/5 px-3">
                  {error}
                </p>
              )}

              <CTAButton
                type="submit"
                disabled={loading}
                className="w-full justify-center"
              >
                {loading ? "Signing in…" : "Sign in"}
              </CTAButton>
            </form>

            <p className="text-center text-xs text-white/30">
              No account?{" "}
              <Link href="/register" className="text-white/60 hover:text-white transition-colors duration-300">
                Register
              </Link>
            </p>
          </Bezel.Inner>
        </Bezel>
      </Reveal>
    </main>
  );
}
