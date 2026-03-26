<script setup lang="ts">
import type { SessionVm } from "../../../lib/api/sessions";

defineProps<{
  canSelectAnySession?: boolean;
  disabled?: boolean;
  loading?: boolean;
  sessions: SessionVm[];
  value: null | string;
}>();

const emit = defineEmits<{
  "update:value": [value: string];
}>();
</script>

<template>
  <fieldset class="fieldset">
    <legend class="fieldset-legend">Session</legend>
    <select
      class="select w-full"
      :value="value ?? ''"
      :disabled="disabled || loading || sessions.length === 0"
      @change="emit('update:value', ($event.target as HTMLSelectElement).value)"
    >
      <option v-if="canSelectAnySession" value="">All Sessions</option>
      <option v-for="session in sessions" :key="session.id" :value="session.id">
        {{ session.name }} ({{ session.startTime.slice(0, 5) }} -
        {{ session.endTime.slice(0, 5) }})
      </option>
    </select>
    <p class="fieldset-label">
      {{
        canSelectAnySession
          ? "Choose a single session or keep all sessions visible."
          : "Your role is limited to its assigned session."
      }}
    </p>
  </fieldset>
</template>
