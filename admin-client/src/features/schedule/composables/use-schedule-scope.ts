import { computed, type ComputedRef, type Ref } from "vue";
import type { SessionVm } from "../../../lib/api/sessions";
import type { UserRole } from "../../../lib/api/types";

export function useScheduleScope(
  sessions: ComputedRef<SessionVm[]>,
  userRole: ComputedRef<UserRole | null> | Ref<UserRole | null>,
) {
  const allowedSessions = computed(() => {
    if (userRole.value === "admin" || !userRole.value) {
      return sessions.value;
    }

    const morningSession = findScopedSession(sessions.value, "morning");
    const afternoonSession = findScopedSession(sessions.value, "afternoon");

    if (userRole.value === "morning_user") {
      return morningSession ? [morningSession] : sessions.value.slice(0, 1);
    }

    if (userRole.value === "afternoon_user") {
      if (afternoonSession) {
        return [afternoonSession];
      }

      return sessions.value.length > 1 ? [sessions.value[1]] : sessions.value;
    }

    return sessions.value;
  });

  const canSelectAnySession = computed(() => userRole.value === "admin");
  const defaultSessionId = computed(() => allowedSessions.value[0]?.id ?? null);

  function isSessionAllowed(sessionId: null | string | undefined) {
    if (!sessionId) {
      return false;
    }

    return allowedSessions.value.some((session) => session.id === sessionId);
  }

  return {
    allowedSessions,
    canSelectAnySession,
    defaultSessionId,
    isSessionAllowed,
  };
}

function findScopedSession(
  sessions: SessionVm[],
  scope: "afternoon" | "morning",
) {
  return sessions.find((session) => session.name.toLowerCase().includes(scope));
}
