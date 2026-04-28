"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, usePathname } from "next/navigation";
import { cn } from "@/lib/cn";
import { clearToken, getToken } from "@/lib/api";

const LINKS = [
  { label: "Play", href: "/play" },
  { label: "Leaderboard", href: "/leaderboard" },
];

/**
 * Floating glass pill nav. Auth-aware: shows Sign in / Get started for guests,
 * Profile / Sign out for authenticated users. Re-reads the token whenever the
 * route changes so login / logout flips immediately without a full reload.
 */
export default function Nav() {
  const [open, setOpen] = useState(false);
  const [authed, setAuthed] = useState<boolean | null>(null);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    setAuthed(!!getToken());
    const onStorage = () => setAuthed(!!getToken());
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, [pathname]);

  function handleSignOut() {
    clearToken();
    setAuthed(false);
    setOpen(false);
    router.push("/");
  }

  // Server-rendered first paint hides auth-specific items to avoid flash.
  const showAuthed = authed === true;
  const showGuest = authed === false;

  return (
    <>
      <header className="fixed top-6 inset-x-0 z-40 flex justify-center pointer-events-none">
        <nav className="pointer-events-auto flex items-center gap-6 px-5 py-2.5 rounded-full backdrop-blur-2xl bg-black/60 ring-1 ring-white/10">
          <Link
            href="/"
            className="font-display text-sm tracking-tight text-white/90 hover:text-white transition-colors duration-500"
          >
            Chess<span className="font-editorial italic">Master</span>
          </Link>

          <div className="hidden md:flex items-center gap-5">
            {LINKS.map((l) => (
              <Link
                key={l.href}
                href={l.href}
                className="text-xs text-white/50 hover:text-white transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] tracking-wide"
              >
                {l.label}
              </Link>
            ))}
            {showAuthed && (
              <Link
                href="/me"
                className="text-xs text-white/50 hover:text-white transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] tracking-wide"
              >
                Profile
              </Link>
            )}
          </div>

          <div className="hidden md:flex items-center gap-2">
            {showGuest && (
              <>
                <Link
                  href="/login"
                  className="text-xs text-white/50 hover:text-white transition-colors duration-500"
                >
                  Sign in
                </Link>
                <Link
                  href="/register"
                  className="text-xs bg-white text-black rounded-full px-3 py-1.5 hover:bg-white/90 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]"
                >
                  Get started
                </Link>
              </>
            )}
            {showAuthed && (
              <button
                onClick={handleSignOut}
                className="text-xs text-white/50 hover:text-rose-400/80 transition-colors duration-500"
              >
                Sign out
              </button>
            )}
          </div>

          <button
            aria-label={open ? "Close menu" : "Open menu"}
            onClick={() => setOpen((v) => !v)}
            className="md:hidden w-7 h-7 flex flex-col justify-center items-center gap-1.5 relative"
          >
            <span
              className={cn(
                "absolute w-4 h-px bg-white transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)]",
                open ? "rotate-45 translate-y-0" : "-translate-y-1"
              )}
            />
            <span
              className={cn(
                "absolute w-4 h-px bg-white transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)]",
                open ? "-rotate-45" : "translate-y-1"
              )}
            />
          </button>
        </nav>
      </header>

      <div
        className={cn(
          "fixed inset-0 z-30 backdrop-blur-3xl bg-black/80 flex flex-col items-center justify-center gap-8 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] md:hidden",
          open ? "opacity-100 pointer-events-auto" : "opacity-0 pointer-events-none"
        )}
      >
        {LINKS.map((l, i) => (
          <Link
            key={l.href}
            href={l.href}
            onClick={() => setOpen(false)}
            style={{ transitionDelay: open ? `${80 + i * 60}ms` : "0ms" }}
            className={cn(
              "font-display text-4xl tracking-tight transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]",
              open ? "translate-y-0 opacity-100" : "translate-y-12 opacity-0"
            )}
          >
            {l.label}
          </Link>
        ))}
        {showAuthed && (
          <>
            <Link
              href="/me"
              onClick={() => setOpen(false)}
              style={{ transitionDelay: open ? "200ms" : "0ms" }}
              className={cn(
                "font-display text-4xl tracking-tight transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]",
                open ? "translate-y-0 opacity-100" : "translate-y-12 opacity-0"
              )}
            >
              Profile
            </Link>
            <button
              onClick={handleSignOut}
              style={{ transitionDelay: open ? "260ms" : "0ms" }}
              className={cn(
                "font-display text-4xl tracking-tight text-rose-400/80 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]",
                open ? "translate-y-0 opacity-100" : "translate-y-12 opacity-0"
              )}
            >
              Sign out
            </button>
          </>
        )}
        {showGuest && (
          <>
            <Link
              href="/login"
              onClick={() => setOpen(false)}
              style={{ transitionDelay: open ? "200ms" : "0ms" }}
              className={cn(
                "font-display text-4xl tracking-tight transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]",
                open ? "translate-y-0 opacity-100" : "translate-y-12 opacity-0"
              )}
            >
              Sign in
            </Link>
            <Link
              href="/register"
              onClick={() => setOpen(false)}
              style={{ transitionDelay: open ? "260ms" : "0ms" }}
              className={cn(
                "font-display text-4xl tracking-tight transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]",
                open ? "translate-y-0 opacity-100" : "translate-y-12 opacity-0"
              )}
            >
              Register
            </Link>
          </>
        )}
      </div>
    </>
  );
}
