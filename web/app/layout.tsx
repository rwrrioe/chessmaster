import type { Metadata } from "next";
import { Inter } from "next/font/google";
import MeshBg from "@/components/MeshBg";
import Nav from "@/components/Nav";
import "./globals.css";

/**
 * Font strategy:
 * - Geist is not available via next/font/google in Next 14.2 (it was added later).
 *   We load it from the Google Fonts CDN via a <link> tag below.
 * - Clash Display + PP Editorial New come from Fontshare CDN via <link>.
 * - Inter is declared here solely to silence the Next font warning; it is not
 *   used in the UI (all font-family rules use Clash Display / Geist / PP Editorial).
 *
 * MeshBg trick: pure CSS radial-gradient on a fixed position:fixed, -z-10,
 * pointer-events-none element — zero repaints during scroll, GPU-composited.
 */
// Inter imported only to satisfy Next font plumbing; never referenced in className.
const _inter = Inter({ subsets: ["latin"], variable: "--font-inter-unused" });

export const metadata: Metadata = {
  title: "ChessMaster Pro",
  description: "Elite chess platform with Gemini-powered coaching.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        {/* Geist sans + mono from Google Fonts CDN */}
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Geist:wght@300;400;500;600;700&family=Geist+Mono:wght@400;500&display=swap"
          rel="stylesheet"
        />
        {/* Clash Display + PP Editorial New from Fontshare CDN */}
        <link rel="preconnect" href="https://api.fontshare.com" />
        <link
          href="https://api.fontshare.com/v2/css?f[]=clash-display@500,600,700&f[]=pp-editorial-new@400i,400&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="font-sans antialiased min-h-[100dvh] bg-[#050505] text-white">
        <MeshBg />
        <Nav />
        {children}
      </body>
    </html>
  );
}
