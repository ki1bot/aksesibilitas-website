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
  title: "Rifqi | AksesCheck ID",
  description:
    "Periksa aksesibilitas website, pahami masalahnya, dan tentukan bagian yang perlu diperbaiki lebih dulu.",
  applicationName: "AksesCheck ID",
  keywords: [
    "aksesibilitas website",
    "WCAG",
    "pemeriksaan website",
    "aksesibilitas digital",
    "accessibility audit",
  ],
  icons: {
    icon: [
      {
        url: "/assets/logoKibot.png",
        type: "image/png",
        sizes: "500x500",
      },
    ],
    shortcut: "/assets/logoKibot.png",
    apple: "/assets/logoKibot.png",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" data-scroll-behavior="smooth" suppressHydrationWarning>
      <body className={`${geistSans.variable} ${geistMono.variable}`}>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
