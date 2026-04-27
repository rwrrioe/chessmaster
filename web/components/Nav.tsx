"use client";
import { useState } from "react";
import Link from "next/link";
import { cn } from "@/lib/cn";

const LINKS = [
  { label: "Play", href: "/play" },
  { label: "Leaderboard", href: "/leaderboard" },
  { label: "My Profile", href: "/me" },
];

/**
 * Floating glass pill nav. Fixed top-6, detached from viewport edge.
 * On mobile: morphing hamburger opens a full-screen overlay with
 * staggered link reveal.
 */
export default function Nav() {
  const [open, setOpen] = useState(false);

  return (
    <>
      {/* Floating pill */}
      <header className="fixed top-6 inset-x-0 z-40 flex justify-center pointer-events-none">
        <nav className="pointer-events-auto flex items-center gap-6 px-5 py-2.5 rounded-full backdrop-blur-2xl bg-black/60 ring-1 ring-white/10">
          {/* Logo */}
          <Link
            href="/"
            className="font-display text-sm tracking-tight text-white/90 hover:text-white transition-colors duration-500"
          >
            Chess<span className="font-editorial italic">Master</span>
          </Link>

          {/* Desktop links */}
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
          </div>

          {/* Auth desktop */}
          <div className="hidden md:flex items-center gap-2">
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
          </div>

          {/* Hamburger — mobile only */}
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

      {/* Mobile overlay */}
      <div
        className={cn(
          "fixed inset-0 z-30 backdrop-blur-3xl bg-black/80 flex flex-col items-center justify-center gap-8 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] md:hidden",
          open ? "opacity-100 pointer-events-auto" : "opacity-0 pointer-events-none"
        )}
      >
        {[...LINKS, { label: "Sign in", href: "/login" }, { label: "Register", href: "/register" }].map(
          (l, i) => (
            <Link
              key={l.href}
              href={l.href}
              onClick={() => setOpen(false)}
              style={{ transitionDelay: open ? `${80 + i * 60}ms` : "0ms" }}
              className={cn(
                "font-display text-4xl tracking-tight transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]",
                open
                  ? "translate-y-0 opacity-100"
                  : "translate-y-12 opacity-0"
              )}
            >
              {l.label}
            </Link>
          )
        )}
      </div>
    </>
  );
}
