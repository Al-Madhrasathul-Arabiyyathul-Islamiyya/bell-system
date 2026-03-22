export const queryKeys = {
  audio: (params?: Record<string, unknown>) => ["audio", params ?? {}] as const,
  audioChecksums: () => ["audio", "checksums"] as const,
  authUser: () => ["auth", "user"] as const,
  currentSchedule: () => ["schedule", "current"] as const,
  currentSession: () => ["sessions", "current"] as const,
  schedule: (params?: Record<string, unknown>) =>
    ["schedule", params ?? {}] as const,
  sessions: (params?: Record<string, unknown>) =>
    ["sessions", params ?? {}] as const,
  systemState: () => ["system", "state"] as const,
  users: (params?: Record<string, unknown>) => ["users", params ?? {}] as const,
} as const;
