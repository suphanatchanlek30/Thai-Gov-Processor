"use client";

import type { PresetId } from "@/lib/types";

interface PresetCardProps {
  id: PresetId;
  label: string;
  description: string;
  selected: boolean;
  onSelect: (id: PresetId) => void;
}

export default function PresetCard({
  id,
  label,
  description,
  selected,
  onSelect,
}: PresetCardProps) {
  return (
    <button
      type="button"
      onClick={() => onSelect(id)}
      aria-pressed={selected}
      className={`flex flex-col gap-1 rounded-lg border p-4 text-left transition-colors ${
        selected
          ? "border-blue-500 bg-blue-50 dark:bg-blue-950/30"
          : "border-gray-300 hover:border-gray-400 dark:border-gray-700 dark:hover:border-gray-600"
      }`}
    >
      <span className="font-medium">{label}</span>
      <span className="text-sm text-gray-500 dark:text-gray-400">{description}</span>
    </button>
  );
}
