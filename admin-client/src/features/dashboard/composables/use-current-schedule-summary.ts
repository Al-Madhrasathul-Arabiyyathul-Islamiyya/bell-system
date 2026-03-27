import { computed, toValue, type MaybeRefOrGetter } from "vue";
import type { CurrentScheduleResponse } from "../../../lib/api/schedule";

export function useCurrentScheduleSummary(
  schedule: MaybeRefOrGetter<CurrentScheduleResponse | null | undefined>,
) {
  const items = computed(() => {
    return [...(toValue(schedule)?.items ?? [])].sort(
      (left, right) => toMinutes(left.time) - toMinutes(right.time),
    );
  });

  const currentItem = computed(() => {
    return items.value.find((item) => item.status === "current") ?? null;
  });

  const upcomingItems = computed(() => {
    return items.value.filter((item) => item.status === "pending");
  });

  const nextItem = computed(() => {
    return upcomingItems.value[0] ?? null;
  });

  return {
    currentItem,
    items,
    nextItem,
    upcomingItems,
  };
}

function toMinutes(time: string) {
  const [hours, minutes] = time.split(":").map((value) => Number(value));

  return (
    (Number.isFinite(hours) ? hours : 0) * 60 +
    (Number.isFinite(minutes) ? minutes : 0)
  );
}
