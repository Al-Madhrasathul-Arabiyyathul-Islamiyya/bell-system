<script setup lang="ts">
import { computed, watch } from "vue";
import { Icon } from "@iconify/vue";
import {
  breakpointsTailwind,
  useBreakpoints,
  useDateFormat,
  useNow,
  useOnline,
  useTitle,
} from "@vueuse/core";
import { RouterLink, RouterView, useRoute, useRouter } from "vue-router";
import { useBellSystemSocket } from "../features/realtime/composables/use-bell-system-socket";
import { appEnv } from "../lib/env";
import { useAppShellStore } from "../stores/app-shell";
import { useAuthStore } from "../stores/auth";
import { usePreferencesStore } from "../stores/preferences";
import { useRealtimeStore } from "../stores/realtime";

const route = useRoute();
const router = useRouter();
const appShellStore = useAppShellStore();
const authStore = useAuthStore();
const preferencesStore = usePreferencesStore();
const realtimeStore = useRealtimeStore();

const online = useOnline();
const now = useNow({ interval: 1_000 });
const formattedTime = useDateFormat(now, "HH:mm:ss");
const breakpoints = useBreakpoints(breakpointsTailwind);
const isDesktop = breakpoints.greaterOrEqual("lg");

const navigationItems = [
  {
    icon: "solar:widget-5-bold-duotone",
    label: "Dashboard",
    to: { name: "dashboard" },
  },
  {
    icon: "solar:calendar-date-bold-duotone",
    label: "Sessions",
    to: { name: "sessions" },
  },
  {
    icon: "solar:calendar-mark-bold-duotone",
    label: "Schedule",
    to: { name: "schedule" },
  },
  {
    icon: "solar:music-library-2-bold-duotone",
    label: "Audio",
    to: { name: "audio" },
  },
  {
    icon: "solar:users-group-rounded-bold-duotone",
    label: "Users",
    to: { name: "users" },
  },
  {
    icon: "solar:settings-bold-duotone",
    label: "System",
    to: { name: "system" },
  },
] as const;

const pageTitle = computed(() => {
  const routeTitle =
    typeof route.meta.title === "string" ? route.meta.title : "Workspace";

  return `${routeTitle} • ${appEnv.appName}`;
});

const drawerOpen = computed(
  () => isDesktop.value || appShellStore.mobileMenuOpen,
);

useBellSystemSocket();
useTitle(pageTitle);

watch(
  () => route.fullPath,
  () => {
    if (!isDesktop.value) {
      appShellStore.closeMobileMenu();
    }
  },
);

watch(isDesktop, (desktop) => {
  if (desktop) {
    appShellStore.closeMobileMenu();
  }
});

function openMobileMenu() {
  if (!isDesktop.value) {
    appShellStore.openMobileMenu();
  }
}

function closeMobileMenu() {
  appShellStore.closeMobileMenu();
}

function handleLogout() {
  authStore.logout();
  router.push({ name: "login" });
}
</script>

<template>
  <div
    class="drawer min-h-screen lg:drawer-open"
    :class="{ 'drawer-open': drawerOpen }"
  >
    <input
      class="drawer-toggle"
      type="checkbox"
      :checked="drawerOpen"
      aria-label="Navigation"
    />

    <div class="drawer-content min-h-screen bg-base-200">
      <header class="border-b border-base-300 bg-base-100/90 backdrop-blur">
        <div class="navbar mx-auto max-w-7xl px-4 sm:px-6">
          <div class="navbar-start gap-3">
            <button
              class="btn btn-ghost btn-square lg:hidden"
              type="button"
              @click="openMobileMenu"
            >
              <Icon icon="solar:hamburger-menu-outline" class="text-xl" />
            </button>
            <div>
              <p
                class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
              >
                Bell System
              </p>
              <h1 class="font-display text-lg font-semibold text-base-content">
                Admin Client
              </h1>
            </div>
          </div>

          <div class="navbar-end gap-2">
            <span
              class="badge badge-outline gap-2 border-base-300 px-3 py-3 text-xs font-medium"
            >
              <span
                class="status"
                :class="online ? 'status-success' : 'status-error'"
              />
              {{ online ? "Online" : "Offline" }}
            </span>
            <span
              class="badge badge-outline gap-2 border-base-300 px-3 py-3 text-xs font-medium"
            >
              <span
                class="status"
                :class="
                  realtimeStore.socketStatus === 'OPEN'
                    ? 'status-success'
                    : realtimeStore.socketStatus === 'CONNECTING'
                      ? 'status-warning'
                      : 'status-error'
                "
              />
              Socket {{ realtimeStore.socketStatus.toLowerCase() }}
            </span>
            <span
              class="badge badge-outline hidden border-base-300 px-3 py-3 text-xs font-medium md:inline-flex"
            >
              {{ formattedTime }}
            </span>
            <div class="dropdown dropdown-end">
              <button class="btn btn-ghost gap-2 px-3" type="button">
                <Icon
                  :icon="
                    preferencesStore.themePreference === 'system'
                      ? 'solar:monitor-bold-duotone'
                      : preferencesStore.resolvedTheme === 'dark'
                        ? 'solar:moon-stars-bold-duotone'
                        : 'solar:sun-bold-duotone'
                  "
                  class="text-xl"
                />
                <span class="hidden text-sm font-medium sm:inline">
                  {{
                    preferencesStore.themePreference === "system"
                      ? "System"
                      : preferencesStore.themePreference === "dark"
                        ? "Dark"
                        : "Light"
                  }}
                </span>
              </button>
              <ul
                class="menu dropdown-content z-10 mt-3 w-40 rounded-box bg-base-100 p-2 shadow"
              >
                <li>
                  <button
                    type="button"
                    @click="preferencesStore.themePreference = 'system'"
                  >
                    System
                  </button>
                </li>
                <li>
                  <button
                    type="button"
                    @click="preferencesStore.themePreference = 'light'"
                  >
                    Light
                  </button>
                </li>
                <li>
                  <button
                    type="button"
                    @click="preferencesStore.themePreference = 'dark'"
                  >
                    Dark
                  </button>
                </li>
              </ul>
            </div>
            <div class="dropdown dropdown-end">
              <button class="btn btn-ghost gap-2 px-3" type="button">
                <span class="hidden text-sm font-medium sm:inline">
                  {{ authStore.username || "admin" }}
                </span>
                <Icon icon="solar:user-circle-bold-duotone" class="text-2xl" />
              </button>
              <ul
                class="menu dropdown-content z-10 mt-3 w-56 rounded-box bg-base-100 p-2 shadow"
              >
                <li
                  class="menu-title text-xs uppercase tracking-[0.2em] text-base-content/50"
                >
                  Current Session
                </li>
                <li>
                  <span
                    class="pointer-events-none text-sm text-base-content/70"
                  >
                    {{ authStore.username || "admin" }}
                  </span>
                </li>
                <li>
                  <button type="button" @click="handleLogout">Sign Out</button>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </header>

      <main
        class="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-7xl flex-col gap-6 px-4 py-6 sm:px-6 lg:py-8"
      >
        <div class="alert alert-info shadow-sm text-sm leading-6">
          API transport, query composables, persisted auth state, and realtime
          socket foundations are now in place. Base URLs are preconfigured as
          <span class="font-semibold">{{ appEnv.apiBaseUrl }}</span>
          and
          <span class="font-semibold">{{ appEnv.wsBaseUrl }}</span
          >. Connected clients:
          <span class="font-semibold">{{
            realtimeStore.connectedClients.length
          }}</span>
        </div>
        <RouterView />
      </main>
    </div>

    <div class="drawer-side z-30">
      <label
        class="drawer-overlay"
        aria-label="Close navigation"
        @click="closeMobileMenu"
      />

      <aside
        class="flex min-h-full w-80 flex-col bg-neutral text-neutral-content"
      >
        <div class="border-b border-white/10 px-6 py-6">
          <div class="flex items-center gap-4">
            <img
              class="size-14 rounded-box bg-white/10 p-2"
              src="/logo.svg"
              alt="Bell System logo"
            />
            <div class="space-y-1">
              <p
                class="text-xs font-semibold uppercase tracking-[0.24em] text-primary-content/70"
              >
                Workspace
              </p>
              <h2 class="font-display text-xl font-semibold">
                Admin Control Surface
              </h2>
              <p class="text-sm text-neutral-content/70">
                Vue, Pinia, Vue Query, DaisyUI
              </p>
            </div>
          </div>
        </div>

        <nav class="flex-1 px-4 py-5">
          <ul class="menu gap-2">
            <li v-for="item in navigationItems" :key="item.label">
              <RouterLink
                class="flex items-center gap-3 rounded-box px-4 py-3 text-sm font-medium"
                active-class="active bg-white/10 text-white"
                :to="item.to"
              >
                <Icon :icon="item.icon" class="text-xl" />
                <span>{{ item.label }}</span>
              </RouterLink>
            </li>
          </ul>
        </nav>

        <div
          class="border-t border-white/10 px-6 py-5 text-sm text-neutral-content/70"
        >
          <p class="font-medium text-neutral-content">
            Next implementation blocks
          </p>
          <p class="mt-2 leading-6">
            Replace placeholder auth with live endpoints, wire dashboard
            queries, and connect Regle forms to the CRUD feature flows.
          </p>
        </div>
      </aside>
    </div>
  </div>
</template>
