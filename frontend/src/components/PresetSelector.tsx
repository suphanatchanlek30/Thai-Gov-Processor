"use client";

import type { PresetId } from "@/lib/types";
import PresetCard from "./PresetCard";

const PRESETS: { id: PresetId; label: string; description: string }[] = [
  { id: "ocsc", label: "OCSC (ก.พ.)", description: "200 × 230 px · ≤ 100 KB" },
  { id: "passport", label: "Passport", description: "500 × 500 px · ≤ 200 KB" },
  {
    id: "teacher",
    label: "Teacher / civil servant",
    description: "150×200 – 300×400 px · ≤ 200 KB",
  },
  { id: "custom", label: "Custom", description: "Set your own width, height and max size" },
];

export interface CustomPresetValues {
  width: string;
  height: string;
  maxKb: string;
}

interface PresetSelectorProps {
  value: PresetId;
  onChange: (id: PresetId) => void;
  custom: CustomPresetValues;
  onCustomChange: (custom: CustomPresetValues) => void;
}

export default function PresetSelector({
  value,
  onChange,
  custom,
  onCustomChange,
}: PresetSelectorProps) {
  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        {PRESETS.map((preset) => (
          <PresetCard
            key={preset.id}
            id={preset.id}
            label={preset.label}
            description={preset.description}
            selected={value === preset.id}
            onSelect={onChange}
          />
        ))}
      </div>

      {value === "custom" && (
        <div className="grid grid-cols-1 gap-3 rounded-lg border border-gray-200 p-4 dark:border-gray-800 sm:grid-cols-3">
          <label className="flex flex-col gap-1 text-sm">
            Width (px)
            <input
              type="number"
              min={1}
              required
              value={custom.width}
              onChange={(e) => onCustomChange({ ...custom, width: e.target.value })}
              className="rounded border border-gray-300 bg-transparent px-2 py-1 dark:border-gray-700"
            />
          </label>
          <label className="flex flex-col gap-1 text-sm">
            Height (px)
            <input
              type="number"
              min={1}
              required
              value={custom.height}
              onChange={(e) => onCustomChange({ ...custom, height: e.target.value })}
              className="rounded border border-gray-300 bg-transparent px-2 py-1 dark:border-gray-700"
            />
          </label>
          <label className="flex flex-col gap-1 text-sm">
            Max size (KB)
            <input
              type="number"
              min={1}
              required
              value={custom.maxKb}
              onChange={(e) => onCustomChange({ ...custom, maxKb: e.target.value })}
              className="rounded border border-gray-300 bg-transparent px-2 py-1 dark:border-gray-700"
            />
          </label>
        </div>
      )}
    </div>
  );
}
