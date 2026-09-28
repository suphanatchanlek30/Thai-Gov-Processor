"use client";

import { useState } from "react";
import Nav from "@/components/Nav";
import DropZone from "@/components/DropZone";
import PresetSelector, { CustomPresetValues } from "@/components/PresetSelector";
import PhotoResultSummary from "@/components/PhotoResultSummary";
import ErrorBanner from "@/components/ErrorBanner";
import { postForm, ApiError } from "@/lib/api";
import { bytesToKb } from "@/lib/format";
import type { PhotoPresetResponse, PresetId } from "@/lib/types";

export default function Home() {
  const [file, setFile] = useState<File | null>(null);
  const [preset, setPreset] = useState<PresetId>("ocsc");
  const [custom, setCustom] = useState<CustomPresetValues>({
    width: "",
    height: "",
    maxKb: "",
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<PhotoPresetResponse | null>(null);

  const handleSubmit = async () => {
    if (!file) {
      setError("Choose a photo first.");
      return;
    }

    setIsSubmitting(true);
    setError(null);
    setResult(null);

    const formData = new FormData();
    formData.append("file", file);
    formData.append("preset", preset);
    if (preset === "custom") {
      formData.append("width", custom.width);
      formData.append("height", custom.height);
      formData.append("max_kb", custom.maxKb);
    }

    try {
      const response = await postForm<PhotoPresetResponse>(
        "/api/v1/photos/preset",
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
          <h1 className="text-2xl font-semibold">Resize a photo</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Pick a preset that matches the agency&rsquo;s requirement, or set your own.
          </p>
        </div>

        <DropZone
          label={file ? file.name : "Drop a photo here, or click to browse"}
          hint="JPG, PNG, HEIC or WEBP, up to 15 MB"
          accept="image/jpeg,image/png,image/heic,image/webp"
          onFilesSelected={(files) => setFile(files[0])}
        />

        <PresetSelector
          value={preset}
          onChange={setPreset}
          custom={custom}
          onCustomChange={setCustom}
        />

        <button
          type="button"
          onClick={handleSubmit}
          disabled={isSubmitting || !file}
          className="w-fit rounded-md bg-blue-600 px-5 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isSubmitting ? "Processing..." : "Convert"}
        </button>

        {error && <ErrorBanner message={error} />}

        {result && file && (
          <PhotoResultSummary
            original={{ name: file.name, sizeKb: bytesToKb(file.size) }}
            result={result}
          />
        )}
      </div>
    </>
  );
}
