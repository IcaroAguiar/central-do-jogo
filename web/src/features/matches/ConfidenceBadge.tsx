import type { ConfidenceLevel } from "../../api/client";
import { confidenceLabel } from "../../lib/availability";

export interface ConfidenceBadgeProps {
  level: ConfidenceLevel;
}

/** Deterministic confidence badge shown next to every broadcast claim
 * (REQ-007). Uses a text label, not color alone, so it stays meaningful
 * without relying on color perception. */
export function ConfidenceBadge({ level }: ConfidenceBadgeProps) {
  return (
    <span className={`confidence-badge confidence-badge--${level}`}>{confidenceLabel(level)}</span>
  );
}
