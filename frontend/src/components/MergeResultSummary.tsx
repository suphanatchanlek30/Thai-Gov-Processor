import type { MergePdfResponse } from "@/lib/types";
import { formatKb } from "@/lib/format";

export default function MergeResultSummary({ result }: { result: MergePdfResponse }) {
  return (
    <div className="flex flex-col gap-4 rounded-lg border border-gray-200 p-4 dark:border-gray-800">
      <div className="flex flex-col gap-1 text-sm">
        <span className="font-medium">{result.filename}</span>
        <span>
          {result.total_pages} pages · {formatKb(result.size_kb)}
        </span>
      </div>
      <a
        href={result.download_url}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex w-fit items-center rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
      >
        Download PDF
      </a>
      <p className="text-xs text-gray-500 dark:text-gray-400">
        Link expires in {Math.round(result.expires_in / 60)} minutes.
      </p>
    </div>
  );
}
