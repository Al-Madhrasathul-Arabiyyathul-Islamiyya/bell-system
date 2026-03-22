import {
  createRouter,
  createWebHistory,
  type RouteLocationNormalized,
  type RouteRecordRaw,
} from "vue-router";
import { pinia } from "./pinia";
import { useAuthStore } from "../stores/auth";

const routes: RouteRecordRaw[] = [
  {
    path: "/login",
    name: "login",
    component: () => import("../pages/LoginPage.vue"),
    meta: {
      layout: "public",
      title: "Login",
    },
  },
  {
    path: "/",
    component: () => import("./AppShell.vue"),
    meta: {
      requiresAuth: true,
    },
    children: [
      {
        path: "",
        redirect: {
          name: "dashboard",
        },
      },
      {
        path: "dashboard",
        name: "dashboard",
        component: () => import("../pages/DashboardPage.vue"),
        meta: {
          title: "Dashboard",
        },
      },
      {
        path: "sessions",
        name: "sessions",
        component: () => import("../pages/SessionsPage.vue"),
        meta: {
          title: "Sessions",
        },
      },
      {
        path: "schedule",
        name: "schedule",
        component: () => import("../pages/SchedulePage.vue"),
        meta: {
          title: "Schedule",
        },
      },
      {
        path: "audio",
        name: "audio",
        component: () => import("../pages/AudioPage.vue"),
        meta: {
          title: "Audio Library",
        },
      },
      {
        path: "users",
        name: "users",
        component: () => import("../pages/UsersPage.vue"),
        meta: {
          title: "Users",
        },
      },
      {
        path: "system",
        name: "system",
        component: () => import("../pages/SystemPage.vue"),
        meta: {
          title: "System",
        },
      },
    ],
  },
  {
    path: "/:pathMatch(.*)*",
    name: "not-found",
    component: () => import("../pages/NotFoundPage.vue"),
    meta: {
      layout: "public",
      title: "Not Found",
    },
  },
];

function requiresAuth(route: RouteLocationNormalized) {
  return route.matched.some((record) => record.meta.requiresAuth);
}

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const authStore = useAuthStore(pinia);

  if (requiresAuth(to) && !authStore.isAuthenticated) {
    return {
      name: "login",
      query: {
        redirect: to.fullPath,
      },
    };
  }

  if (to.name === "login" && authStore.isAuthenticated) {
    return typeof to.query.redirect === "string"
      ? to.query.redirect
      : "/dashboard";
  }

  return true;
});
