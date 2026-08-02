import Link from "next/link";
import { ScanSearch } from "lucide-react";

import { cn } from "@/lib/utils";

export function Brand({
  compact = false,
  href = "/",
  className,
}: {
  compact?: boolean;
  href?: string;
  className?: string;
}) {
  return (
    <Link
      href={href}
      className={cn(
        "inline-flex items-center gap-3 rounded-xl focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/25",
        className,
      )}
      aria-label="AksesCheck ID"
    >
      <span className="flex size-10 items-center justify-center rounded-xl bg-blue-600 text-white shadow-lg shadow-blue-600/20">
        <ScanSearch className="size-5" aria-hidden="true" />
      </span>

      {!compact && (
        <span className="flex flex-col">
          <span className="text-base font-bold tracking-tight text-slate-950 dark:text-white">
            AksesCheck
          </span>
          <span className="text-[10px] font-bold uppercase tracking-[0.24em] text-blue-600 dark:text-blue-400">
            Indonesia
          </span>
        </span>
      )}
    </Link>
  );
}
