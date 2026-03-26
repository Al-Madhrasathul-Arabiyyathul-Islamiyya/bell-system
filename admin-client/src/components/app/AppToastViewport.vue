<script setup lang="ts">
import { computed } from "vue";
import { Icon } from "@iconify/vue";
import { useRoute } from "vue-router";
import { useToastStore, type ToastTone } from "../../stores/toast";

const route = useRoute();
const toastStore = useToastStore();

const positionClass = computed(() => {
  return route.meta.layout === "public" ? "top-4" : "top-20";
});

function toneClass(tone: ToastTone) {
  if (tone === "success") {
    return "alert-success";
  }

  if (tone === "warning") {
    return "alert-warning";
  }

  if (tone === "error") {
    return "alert-error";
  }

  return "alert-info";
}

function toneIcon(tone: ToastTone) {
  if (tone === "success") {
    return "solar:check-circle-bold-duotone";
  }

  if (tone === "warning") {
    return "solar:danger-triangle-bold-duotone";
  }

  if (tone === "error") {
    return "solar:close-circle-bold-duotone";
  }

  return "solar:info-circle-bold-duotone";
}
</script>

<template>
  <div
    class="pointer-events-none fixed right-4 z-50 flex w-[min(26rem,calc(100vw-2rem))] flex-col gap-3"
    :class="positionClass"
  >
    <TransitionGroup name="toast">
      <div
        v-for="item in toastStore.items"
        :key="item.id"
        class="pointer-events-auto"
      >
        <div class="alert shadow-lg" :class="toneClass(item.tone)">
          <Icon :icon="toneIcon(item.tone)" class="text-2xl" />
          <div class="min-w-0">
            <h2 class="text-base font-semibold leading-6">{{ item.title }}</h2>
            <p class="text-base leading-6 opacity-90">{{ item.detail }}</p>
          </div>
          <button
            class="btn btn-ghost btn-sm btn-circle"
            type="button"
            aria-label="Dismiss notification"
            @click="toastStore.dismiss(item.id)"
          >
            <Icon icon="solar:close-line-duotone" class="text-xl" />
          </button>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>
