import React from "react";
import { cn } from "@/lib/cn";

interface BezelProps {
  children: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

interface BezelInnerProps {
  children: React.ReactNode;
  className?: string;
}

/**
 * Double-Bezel (Doppelrand) component.
 * Usage: <Bezel><Bezel.Inner>content</Bezel.Inner></Bezel>
 *
 * Outer shell: subtle bg + hairline ring + p-1.5 + large radius
 * Inner core:  own bg + inset highlight + concentric smaller radius
 */
function Bezel({ children, className, style }: BezelProps) {
  return (
    <div
      style={style}
      className={cn(
        "bg-white/5 ring-1 ring-white/10 p-1.5 rounded-[2rem]",
        className
      )}
    >
      {children}
    </div>
  );
}

function BezelInner({ children, className }: BezelInnerProps) {
  return (
    <div
      className={cn(
        "rounded-[calc(2rem-0.375rem)] bg-[#0a0a0a] shadow-[inset_0_1px_1px_rgba(255,255,255,0.15)]",
        className
      )}
    >
      {children}
    </div>
  );
}

Bezel.Inner = BezelInner;

export default Bezel;
