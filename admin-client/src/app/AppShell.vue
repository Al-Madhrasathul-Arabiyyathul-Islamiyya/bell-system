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
import { useLogoutMutation } from "../features/auth/composables/use-auth";
import { useBellSystemSocket } from "../features/realtime/composables/use-bell-system-socket";
import { appEnv } from "../lib/env";
import { useAppShellStore } from "../stores/app-shell";
import { useAuthStore } from "../stores/auth";
import { usePreferencesStore } from "../stores/preferences";
import { useRealtimeStore } from "../stores/realtime";
import { useToastStore } from "../stores/toast";

const route = useRoute();
const router = useRouter();
const appShellStore = useAppShellStore();
const authStore = useAuthStore();
const toastStore = useToastStore();
const preferencesStore = usePreferencesStore();
const realtimeStore = useRealtimeStore();
const logoutMutation = useLogoutMutation();

const online = useOnline();
const now = useNow({ interval: 1_000 });
const formattedDate = useDateFormat(now, "ddd, DD MMM YYYY");
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
const sidebarCollapsed = computed(
  () => isDesktop.value && appShellStore.sidebarCollapsed,
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

function toggleSidebar() {
  if (isDesktop.value) {
    appShellStore.toggleSidebarCollapsed();
    return;
  }

  if (appShellStore.mobileMenuOpen) {
    appShellStore.closeMobileMenu();
    return;
  }

  appShellStore.openMobileMenu();
}

function closeMobileMenu() {
  appShellStore.closeMobileMenu();
}

async function handleLogout() {
  try {
    await logoutMutation.mutateAsync();
  } catch {
    // Clear local auth and return to login even if the server logout call fails.
  }

  toastStore.enqueue({
    detail: "You have been signed out of the admin client.",
    title: "Signed Out",
    tone: "success",
  });
  await router.push({ name: "login" });
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
        <div class="navbar w-full px-4 sm:px-6">
          <div class="navbar-start gap-3">
            <button
              class="btn btn-ghost btn-square"
              type="button"
              @click="toggleSidebar"
            >
              <Icon
                :icon="
                  isDesktop
                    ? sidebarCollapsed
                      ? 'solar:alt-arrow-right-bold-duotone'
                      : 'solar:alt-arrow-left-bold-duotone'
                    : 'solar:hamburger-menu-outline'
                "
                class="text-xl"
              />
            </button>
            <div class="flex items-center gap-3">
              <img
                class="size-11 rounded-box bg-primary/10 p-2"
                src="/logo.svg"
                alt="Bell System logo"
              />
              <div class="min-w-0">
                <h1
                  class="font-display text-lg font-semibold text-base-content"
                >
                  Arabiyya Bell System
                </h1>
                <p class="truncate text-sm text-base-content/60">
                  Admin Client
                </p>
              </div>
            </div>
            <div class="hidden flex-col leading-tight md:flex">
              <p class="text-sm font-semibold text-base-content/70">
                {{ formattedDate }}
                <span
                  class="ml-2 text-xs font-medium uppercase tracking-[0.18em]"
                >
                  Local Time
                </span>
              </p>
              <p class="font-display text-2xl font-semibold text-base-content">
                {{ formattedTime }}
                <span
                  class="ml-2 font-sans text-sm font-medium text-base-content/60"
                >
                  Local Time
                </span>
              </p>
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
        class="flex min-h-[calc(100vh-4rem)] w-full flex-col gap-6 px-4 py-6 sm:px-6 lg:py-8"
      >
        <div class="alert alert-info shadow-sm text-sm leading-6">
          API transport, query composables, persisted auth state, and realtime
          socket foundations are now in place. Base URLs are preconfigured as
          <span class="font-semibold">{{ appEnv.apiBaseUrl }}</span>
          and
          <span class="font-semibold">{{ appEnv.wsBaseUrl }}</span
          >Connected clients:
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
        class="flex min-h-full w-80 flex-col border-r border-base-300 bg-base-100 text-base-content transition-[width] duration-200 ease-out"
        :class="{
          'lg:w-24': sidebarCollapsed,
          'lg:w-80': !sidebarCollapsed,
        }"
      >
        <div class="border-b border-base-300 px-6 py-6">
          <div
            class="flex items-center gap-4"
            :class="{ 'justify-center': sidebarCollapsed }"
          >
            <img
              class="size-14 rounded-box bg-primary/10 p-2"
              src="/logo.svg"
              alt="Bell System logo"
            />
          </div>
        </div>

        <nav class="flex-1 px-4 py-5">
          <ul class="menu w-full gap-2">
            <li v-for="item in navigationItems" :key="item.label">
              <RouterLink
                class="flex w-full items-center gap-3 rounded-box px-4 py-3 text-sm font-medium"
                :class="{ 'justify-center px-3': sidebarCollapsed }"
                active-class="active"
                :to="item.to"
                :title="sidebarCollapsed ? item.label : undefined"
              >
                <Icon :icon="item.icon" class="text-xl" />
                <span v-if="!sidebarCollapsed">{{ item.label }}</span>
              </RouterLink>
            </li>
          </ul>
        </nav>
      </aside>
    </div>
  </div>
</template>
