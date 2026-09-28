export function formatKb(kb: number): string {
  if (kb >= 1024) return `${(kb / 1024).toFixed(2)} MB`;
  return `${kb.toFixed(1)} KB`;
}

export function bytesToKb(bytes: number): number {
  return bytes / 1024;
}
