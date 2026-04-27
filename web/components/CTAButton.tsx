"use client";
import { ArrowUpRight } from "phosphor-react";
import { cn } from "@/lib/cn";

interface CTAButtonProps {
  children: React.ReactNode;
  onClick?: () => void;
  href?: string;
  showIcon?: boolean;
  className?: string;
  variant?: "primary" | "ghost";
  type?: "button" | "submit";
  disabled?: boolean;
}

/**
 * Primary pill CTA with optional Button-in-Button trailing icon.
 * Hover: scale + icon translates diagonally for kinetic tension.
 */
export default function CTAButton({
  children,
  onClick,
  href,
  showIcon = true,
  className,
  variant = "primary",
  type = "button",
  disabled,
}: CTAButtonProps) {
  const base = cn(
    "group inline-flex items-center gap-3 rounded-full px-5 py-2.5 font-medium text-sm",
    "transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]",
    "active:scale-[0.98]",
    variant === "primary"
      ? "bg-white text-black hover:bg-white/90"
      : "ring-1 ring-white/10 bg-white/5 text-white hover:bg-white/10",
    disabled && "opacity-40 pointer-events-none",
    className
  );

  const icon = showIcon && (
    <span className="w-7 h-7 rounded-full bg-black/10 flex items-center justify-center transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:scale-105">
      <ArrowUpRight size={13} weight="bold" />
    </span>
  );

  if (href) {
    return (
      <a href={href} className={base}>
        <span>{children}</span>
        {icon}
      </a>
    );
  }

  return (
    <button type={type} onClick={onClick} disabled={disabled} className={base}>
      <span>{children}</span>
      {icon}
    </button>
  );
}
