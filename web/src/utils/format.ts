export function clampPercent(value: number): number {
  if (!Number.isFinite(value)) return 0;
  if (value < 0) return 0;
  if (value > 100) return 100;
  return value;
}

export function formatPercent(value?: number | null, decimals = 1): string {
  if (value === null || value === undefined || !Number.isFinite(value)) return '—';
  const clamped = clampPercent(value);
  const factor = Math.pow(10, decimals);
  const rounded = Math.round(clamped * factor) / factor;
  if (Number.isInteger(rounded)) return `${rounded}%`;
  return `${rounded.toFixed(decimals)}%`;
}

export function formatBytes(value?: number | null, decimals = 1): string {
  if (value === null || value === undefined || !Number.isFinite(value) || value < 0) return '—';
  const abs = Math.abs(value);
  const kb = 1e3;
  const mb = 1e6;
  const gb = 1e9;
  const tb = 1e12;

  if (abs >= tb) return `${(value / tb).toFixed(decimals)} TB`;
  if (abs >= gb) return `${(value / gb).toFixed(decimals)} GB`;
  if (abs >= mb) return `${(value / mb).toFixed(decimals)} MB`;
  if (abs >= kb) return `${(value / kb).toFixed(decimals)} KB`;
  return `${Math.round(value)} B`;
}

export function formatUptimeSeconds(value?: number | null): string {
  if (value === null || value === undefined || !Number.isFinite(value) || value < 0) return '—';
  const total = Math.floor(value);
  const days = Math.floor(total / 86400);
  const hours = Math.floor((total % 86400) / 3600);
  const minutes = Math.floor((total % 3600) / 60);

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${Math.max(0, minutes)}m`;
}
