export type PresetId = "ocsc" | "passport" | "teacher" | "custom";

export interface PhotoPresetResponse {
  filename: string;
  width: number;
  height: number;
  size_kb: number;
  quality: number;
  download_url: string;
  expires_in: number;
}

export interface MergePdfResponse {
  filename: string;
  total_pages: number;
  size_kb: number;
  download_url: string;
  expires_in: number;
}

export interface ApiErrorResponse {
  error: string;
}
