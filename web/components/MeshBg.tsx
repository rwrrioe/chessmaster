/**
 * MeshBg renders a fixed full-viewport background with two large radial
 * gradient orbs (purple + emerald) at offset positions.
 * It sits at -z-10, pointer-events-none, so it never intercepts clicks.
 *
 * Trick: pure CSS radial-gradient, no canvas or SVG — zero runtime cost.
 * The gradients are baked into a single background shorthand so there is
 * only one element in the paint tree.
 */
export default function MeshBg() {
  return (
    <div
      aria-hidden
      className="fixed inset-0 -z-10 pointer-events-none"
      style={{
        background: `
          radial-gradient(ellipse 60% 50% at 15% 20%, rgba(91,33,182,0.06) 0%, transparent 70%),
          radial-gradient(ellipse 50% 40% at 85% 75%, rgba(5,150,105,0.04) 0%, transparent 70%),
          #050505
        `,
      }}
    />
  );
}
