import type { App as VueApp } from "vue";
import { VueQueryPlugin } from "@tanstack/vue-query";
import { pinia } from "./pinia";
import { router } from "./router";
import { queryClient } from "./query-client";

export function installAppProviders(app: VueApp<Element>) {
  app.use(pinia);
  app.use(router);
  app.use(VueQueryPlugin, {
    queryClient,
  });
}
