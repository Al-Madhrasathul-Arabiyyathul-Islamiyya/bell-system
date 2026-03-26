<script setup lang="ts">
import { computed } from "vue";
import { Icon } from "@iconify/vue";
import { usePreferencesStore } from "../../stores/preferences";
import type { ThemePreference } from "../../stores/preferences";
import { useDropdown } from "../../composables/useDropdown";

withDefaults(
  defineProps<{
    iconOnly?: boolean;
  }>(),
  {
    iconOnly: false,
  },
);

const preferencesStore = usePreferencesStore();
const { open, toggle, close } = useDropdown();

const themeLabel = computed(() => {
  if (preferencesStore.themePreference === "system") return "System";
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
  close();
}
</script>

<template>
  <div class="dropdown dropdown-end" :class="{ 'dropdown-open': open }">
    <div
      ref="triggerRef"
      tabindex="0"
      role="button"
      class="btn btn-ghost"
      :class="iconOnly ? 'btn-circle' : 'gap-2 px-3'"
      aria-label="Theme switcher"
      @click="toggle"
    >
      <Icon :icon="themeIcon" class="text-xl" />
      <span v-if="!iconOnly" class="hidden text-sm font-medium sm:inline">
        {{ themeLabel }}
      </span>
    </div>

    <ul
      v-show="open"
      ref="dropdownRef"
      class="menu absolute right-0 z-70 mt-3 w-40 rounded-box border border-base-300 bg-base-100 p-2 shadow-xl"
      @click.stop
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
