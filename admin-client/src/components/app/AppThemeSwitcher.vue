<script setup lang="ts">
import { computed, useTemplateRef } from "vue";
import { Icon } from "@iconify/vue";
import { usePreferencesStore } from "../../stores/preferences";
import type { ThemePreference } from "../../stores/preferences";

withDefaults(
  defineProps<{
    iconOnly?: boolean;
  }>(),
  {
    iconOnly: false,
  },
);

const preferencesStore = usePreferencesStore();
const trigger = useTemplateRef<HTMLDivElement>("trigger");

const themeLabel = computed(() => {
  if (preferencesStore.themePreference === "system") {
    return "System";
  }

  return preferencesStore.themePreference === "dark" ? "Dark" : "Light";
});

const themeIcon = computed(() => {
  if (preferencesStore.themePreference === "system") {
    return "solar:monitor-bold-duotone";
  }

  return preferencesStore.resolvedTheme === "dark"
    ? "solar:moon-stars-bold-duotone"
    : "solar:sun-bold-duotone";
});

function setTheme(theme: ThemePreference) {
  preferencesStore.themePreference = theme;
  trigger.value?.blur();
}
</script>

<template>
  <div class="dropdown dropdown-end">
    <div
      ref="trigger"
      tabindex="0"
      role="button"
      class="btn btn-ghost"
      :class="iconOnly ? 'btn-circle' : 'gap-2 px-3'"
      aria-label="Theme switcher"
    >
      <Icon :icon="themeIcon" class="text-xl" />
      <span v-if="!iconOnly" class="hidden text-sm font-medium sm:inline">
        {{ themeLabel }}
      </span>
    </div>
    <ul
      class="menu dropdown-content z-[70] mt-3 w-40 rounded-box border border-base-300 bg-base-100 p-2 shadow-xl"
      tabindex="-1"
    >
      <li>
        <button type="button" @click="setTheme('system')">System</button>
      </li>
      <li>
        <button type="button" @click="setTheme('light')">Light</button>
      </li>
      <li>
        <button type="button" @click="setTheme('dark')">Dark</button>
      </li>
    </ul>
  </div>
</template>
