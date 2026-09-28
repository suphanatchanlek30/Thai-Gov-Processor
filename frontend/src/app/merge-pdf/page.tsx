"use client";

import { useState } from "react";
import Nav from "@/components/Nav";
import DropZone from "@/components/DropZone";
import MergeResultSummary from "@/components/MergeResultSummary";
import ErrorBanner from "@/components/ErrorBanner";
import { postForm, ApiError } from "@/lib/api";
import type { MergePdfResponse } from "@/lib/types";

export default function MergePdfPage() {
  const [files, setFiles] = useState<File[]>([]);
  const [targetMaxKb, setTargetMaxKb] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<MergePdfResponse | null>(null);

  const handleSubmit = async () => {
    if (files.length === 0) {
      setError("Choose at least one file.");
      return;
    }

    setIsSubmitting(true);
    setError(null);
    setResult(null);

    const formData = new FormData();
    files.forEach((file) => formData.append("files[]", file));
    // Blank means "let the backend apply its own default (500 KB)".
    if (targetMaxKb.trim() !== "") {
      formData.append("target_max_kb", targetMaxKb);
    }

    try {
      const response = await postForm<MergePdfResponse>(
        "/api/v1/documents/merge-pdf",
        formData
      );
      setResult(response);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <>
      <Nav />
      <div className="flex flex-col gap-6">
        <div>
          <h1 className="text-2xl font-semibold">Merge to PDF</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Combine images and/or PDFs into a single PDF, in the order you add them.
          </p>
        </div>

        <DropZone
          label={
            files.length > 0
              ? `${files.length} file${files.length > 1 ? "s" : ""} selected`
              : "Drop images or PDFs here, or click to browse"
          }
          hint="Images and PDFs, multiple files allowed"
          accept="image/*,application/pdf"
          multiple
          onFilesSelected={(newFiles) => setFiles((prev) => [...prev, ...newFiles])}
        />

        {files.length > 0 && (
          <ul className="flex flex-col gap-1 text-sm">
            {files.map((file, index) => (
              <li key={`${file.name}-${index}`} className="flex items-center justify-between">
                <span>
                  {index + 1}. {file.name}
                </span>
                <button
                  type="button"
                  onClick={() => setFiles((prev) => prev.filter((_, i) => i !== index))}
                  className="text-gray-500 hover:text-red-600 dark:text-gray-400"
                >
                  Remove
                </button>
              </li>
            ))}
          </ul>
        )}

        <label className="flex w-fit flex-col gap-1 text-sm">
          Target max size (KB) — optional, defaults to 500
          <input
            type="number"
            min={1}
            value={targetMaxKb}
            onChange={(e) => setTargetMaxKb(e.target.value)}
            placeholder="500"
            className="rounded border border-gray-300 bg-transparent px-2 py-1 dark:border-gray-700"
          />
        </label>

        <button
          type="button"
          onClick={handleSubmit}
          disabled={isSubmitting || files.length === 0}
          className="w-fit rounded-md bg-blue-600 px-5 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isSubmitting ? "Merging..." : "Merge"}
        </button>

        {error && <ErrorBanner message={error} />}

        {result && <MergeResultSummary result={result} />}
      </div>
    </>
  );
}
