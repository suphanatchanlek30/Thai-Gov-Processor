import type { PhotoPresetResponse } from "@/lib/types";
import { formatKb } from "@/lib/format";

interface PhotoResultSummaryProps {
  original: { name: string; sizeKb: number };
  result: PhotoPresetResponse;
}

export default function PhotoResultSummary({ original, result }: PhotoResultSummaryProps) {
  return (
    <div className="flex flex-col gap-4 rounded-lg border border-gray-200 p-4 dark:border-gray-800">
      <div className="grid grid-cols-2 gap-4 text-sm">
        <div className="flex flex-col gap-1">
          <span className="text-gray-500 dark:text-gray-400">Before</span>
          <span className="font-medium">{original.name}</span>
          <span>{formatKb(original.sizeKb)}</span>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-gray-500 dark:text-gray-400">After</span>
          <span className="font-medium">{result.filename}</span>
          <span>{formatKb(result.size_kb)}</span>
          <span>
            {result.width} × {result.height} px · quality {result.quality}
          </span>
        </div>
      </div>
      <a
        href={result.download_url}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex w-fit items-center rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
      >
        Download result
      </a>
      <p className="text-xs text-gray-500 dark:text-gray-400">
        Link expires in {Math.round(result.expires_in / 60)} minutes.
      </p>
    </div>
  );
}
