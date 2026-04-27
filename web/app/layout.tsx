import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "ChessMaster Pro",
  description: "Elite chess platform with Gemini-powered coaching.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="font-sans antialiased min-h-[100dvh]">{children}</body>
    </html>
  );
}
