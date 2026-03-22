import { apiJson, apiJsonApiWrite } from "./client";

export type SystemStateVm = {
  lastUpdated: string;
  state: "active" | "paused";
};

export type UpdateSystemStatePayload = {
  data: {
    attributes: {
      state: SystemStateVm["state"];
    };
    type: "system-state";
  };
};

export async function getSystemState() {
  return apiJson<SystemStateVm>("/system/state");
}

export async function updateSystemState(payload: UpdateSystemStatePayload) {
  return apiJsonApiWrite<void>("/system/state", {
    body: payload,
    method: "POST" as const,
  });
}

export async function cancelNextBell() {
  return apiJson<void>("/system/cancel-next-bell", {
    method: "POST" as const,
    requestType: "json",
  });
}
