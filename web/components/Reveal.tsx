"use client";
import { useReveal } from "@/lib/useReveal";
import { cn } from "@/lib/cn";

interface RevealProps {
  children: React.ReactNode;
  className?: string;
  delay?: number;
}

/**
 * Reveal wraps children in a div that fades up + unblurs once it
 * enters the viewport via IntersectionObserver.
 */
export default function Reveal({ children, className, delay = 0 }: RevealProps) {
  const [ref, visible] = useReveal();

  return (
    <div
      ref={ref as React.Ref<HTMLDivElement>}
      style={{ transitionDelay: `${delay}ms` }}
      className={cn(
        "transition-all duration-[800ms] ease-[cubic-bezier(0.32,0.72,0,1)]",
        visible
          ? "opacity-100 translate-y-0 blur-0"
          : "opacity-0 translate-y-16 blur-md",
        className
      )}
    >
      {children}
    </div>
  );
}
