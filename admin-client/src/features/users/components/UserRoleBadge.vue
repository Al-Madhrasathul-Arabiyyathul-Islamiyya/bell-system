<script setup lang="ts">
import { computed } from "vue";
import type { UserRole } from "../../../lib/api/types";

const props = defineProps<{
  role: UserRole;
}>();

const label = computed(() => {
  if (props.role === "morning_user") {
    return "Morning User";
  }

  if (props.role === "afternoon_user") {
    return "Afternoon User";
  }

  return "Admin";
});

const badgeClass = computed(() => {
  if (props.role === "admin") {
    return "badge-primary";
  }

  if (props.role === "morning_user") {
    return "badge-secondary";
  }

  return "badge-accent";
});
</script>

<template>
  <span class="badge badge-soft font-medium" :class="badgeClass">
    {{ label }}
  </span>
</template>
