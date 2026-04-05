import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Smile Dental Clinic — Newmarket, Ontario",
  description:
    "Smile Dental Clinic offers comprehensive dental care in Newmarket, ON. Book your appointment today for checkups, cleanings, whitening, implants, Invisalign, and more.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={inter.className}>{children}</body>
    </html>
  );
}
