"use client";
import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { register, setToken } from "@/lib/api";
import Bezel from "@/components/Bezel";
import EyebrowTag from "@/components/EyebrowTag";
import CTAButton from "@/components/CTAButton";
import Reveal from "@/components/Reveal";

export default function RegisterPage() {
  const router = useRouter();
  const [form, setForm] = useState({ email: "", username: "", password: "", city: "" });
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  function update(field: string) {
    return (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((prev) => ({ ...prev, [field]: e.target.value }));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const { token } = await register(form.email, form.username, form.password, form.city);
      setToken(token);
      router.push("/play");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  const fields: Array<{ key: keyof typeof form; label: string; type: string; placeholder: string }> = [
    { key: "email", label: "Email", type: "email", placeholder: "you@example.com" },
    { key: "username", label: "Username", type: "text", placeholder: "grandmaster_99" },
    { key: "password", label: "Password", type: "password", placeholder: "••••••••" },
    { key: "city", label: "City (optional)", type: "text", placeholder: "Almaty" },
  ];

  return (
    <main className="min-h-[100dvh] flex items-center justify-center px-4 py-28">
      <Reveal className="w-full max-w-sm">
        <Bezel>
          <Bezel.Inner className="p-8 space-y-6">
            <div className="text-center space-y-3">
              <EyebrowTag>Join the platform</EyebrowTag>
              <h1 className="font-display text-3xl tracking-tight">Create account</h1>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              {fields.map(({ key, label, type, placeholder }) => (
                <div key={key} className="space-y-1.5">
                  <label className="text-xs text-white/40 tracking-wide uppercase">
                    {label}
                  </label>
                  <input
                    type={type}
                    required={key !== "city"}
                    value={form[key]}
                    onChange={update(key)}
                    className="w-full rounded-2xl bg-white/5 ring-1 ring-white/10 px-4 py-3 text-sm text-white placeholder-white/20 outline-none focus:ring-white/30 transition-all duration-500"
                    placeholder={placeholder}
                  />
                </div>
              ))}

              {error && (
                <p className="text-xs text-rose-400/80 text-center py-2 ring-1 ring-rose-500/20 rounded-xl bg-rose-500/5 px-3">
                  {error}
                </p>
              )}

              <CTAButton type="submit" disabled={loading} className="w-full justify-center">
                {loading ? "Creating account…" : "Get started"}
              </CTAButton>
            </form>

            <p className="text-center text-xs text-white/30">
              Have an account?{" "}
              <Link href="/login" className="text-white/60 hover:text-white transition-colors duration-300">
                Sign in
              </Link>
            </p>
          </Bezel.Inner>
        </Bezel>
      </Reveal>
    </main>
  );
}
