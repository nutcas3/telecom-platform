"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Users,
  LayoutDashboard,
  Activity,
  CreditCard,
  Shield,
  Settings,
  Radio,
  HeartPulse,
  TrendingDown,
  AlertTriangle,
  BarChart3,
  DollarSign,
  Wrench,
} from "lucide-react";
import { cn } from "@/lib/utils";

const navItems = [
  { href: "/", label: "Dashboard", icon: LayoutDashboard },
  { href: "/subscribers", label: "Subscribers", icon: Users },
  { href: "/usage", label: "Usage & Billing", icon: Activity },
  { href: "/payments", label: "Payments", icon: CreditCard },
  { href: "/esim", label: "eSIM Profiles", icon: Radio },
  { href: "/analytics", label: "Analytics", icon: BarChart3 },
  { href: "/churn", label: "Churn Analysis", icon: TrendingDown },
  { href: "/fraud", label: "Fraud Detection", icon: AlertTriangle },
  { href: "/pricing", label: "Pricing", icon: DollarSign },
  { href: "/maintenance", label: "Maintenance", icon: Wrench },
  { href: "/health", label: "System Health", icon: HeartPulse },
  { href: "/chaos", label: "Chaos Engineering", icon: Shield },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden lg:flex flex-col w-64 border-r border-border bg-sidebar text-sidebar-foreground">
      <div className="flex items-center gap-2 px-6 py-5 border-b border-sidebar-border">
        <Radio className="size-6 text-sidebar-primary" />
        <span className="text-lg font-semibold tracking-tight">Telecom Admin</span>
      </div>

      <nav className="flex-1 py-4 px-3 space-y-1">
        {navItems.map((item) => {
          const active =
            item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                active
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-accent-foreground"
              )}
            >
              <item.icon className="size-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>

    </aside>
  );
}
