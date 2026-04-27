import { cn } from "@/lib/cn";

interface EyebrowTagProps {
  children: React.ReactNode;
  className?: string;
}

/**
 * Small uppercase pill badge to precede major headings.
 */
export default function EyebrowTag({ children, className }: EyebrowTagProps) {
  return (
    <span
      className={cn(
        "inline-block rounded-full px-3 py-1 text-[10px] uppercase tracking-[0.2em] font-medium ring-1 ring-white/10 bg-white/5 text-white/60",
        className
      )}
    >
      {children}
    </span>
  );
}
