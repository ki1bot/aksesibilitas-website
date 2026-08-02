import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";

import { Providers } from "@/components/providers";

import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    default: "AksesCheck ID",
    template: "%s | AksesCheck ID",
  },
  description:
    "Platform pemeriksaan aksesibilitas website menggunakan Go, chromedp, axe-core, dan WCAG 2.2.",
  applicationName: "AksesCheck ID",
  keywords: [
    "aksesibilitas",
    "WCAG",
    "axe-core",
    "website scanner",
    "accessibility audit",
  ],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="id"
      suppressHydrationWarning
      className={`${geistSans.variable} ${geistMono.variable}`}
    >
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
