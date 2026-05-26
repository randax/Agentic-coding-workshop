import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Link from "next/link";
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
  title: "ISP CRM",
  description: "Internal CRM for an ISP — customers, products, and support cases.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <header className="border-b border-gray-200 bg-white">
          <nav className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4">
            <span className="font-semibold text-gray-900">ISP CRM</span>
            <Link href="/" className="text-sm text-gray-600 hover:text-gray-900">
              Customers
            </Link>
            <Link
              href="/products"
              className="text-sm text-gray-600 hover:text-gray-900"
            >
              Products
            </Link>
          </nav>
        </header>
        {children}
      </body>
    </html>
  );
}
