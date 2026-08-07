"use client";

import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";

export function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme();

  function toggleTheme() {
    setTheme(resolvedTheme === "dark" ? "light" : "dark");
  }

  return (
    <button
      type="button"
      onClick={toggleTheme}
      className="grid size-10 shrink-0 place-items-center rounded-xl border border-slate-200 bg-white text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-100 dark:border-white/10 dark:bg-white/5 dark:text-slate-200 dark:hover:border-white/20 dark:hover:bg-white/10"
      aria-label="Ganti tema"
      title="Ganti tema"
    >
      <Moon className="size-4.5 dark:hidden" aria-hidden="true" />
      <Sun className="hidden size-4.5 dark:block" aria-hidden="true" />
    </button>
  );
}
