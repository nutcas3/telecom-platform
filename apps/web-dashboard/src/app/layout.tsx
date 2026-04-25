import type { Metadata } from "next";
import { Geist, Geist_Mono, Inter } from "next/font/google";
import "./globals.css";
import { cn } from "@/lib/utils";
import { Sidebar } from "@/components/sidebar";
import { Providers } from "./providers";
import { ServiceWorkerRegistration } from "@/components/service-worker-registration";

const inter = Inter({ subsets: ["latin"], variable: "--font-sans" });

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Telecom Admin Dashboard",
  description: "Sovereign Telecom-as-a-Service Platform Admin Dashboard",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={cn("h-full", "antialiased", geistSans.variable, geistMono.variable, "font-sans", inter.variable)}
    >
      <body className="min-h-full flex">
        <Providers>
          <ServiceWorkerRegistration />
          <Sidebar />
          <main className="flex-1 overflow-auto">{children}</main>
        </Providers>
      </body>
    </html>
  );
}
