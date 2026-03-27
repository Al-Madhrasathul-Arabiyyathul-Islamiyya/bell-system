import { adaptJsonApiResource } from "../jsonapi/adapters";
import { apiJsonApiDocument, apiJsonApiWrite } from "./client";

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
  const document = await apiJsonApiDocument<{
    lastUpdated: string;
    state: SystemStateVm["state"];
  }>("/system/state");

  return adaptJsonApiResource(document, (resource) => ({
    lastUpdated: resource.attributes.lastUpdated,
    state: resource.attributes.state,
  }));
}

export async function updateSystemState(payload: UpdateSystemStatePayload) {
  return apiJsonApiWrite<void>("/system/state", {
    body: payload,
    method: "POST" as const,
  });
}

export async function cancelNextBell() {
  return apiJsonApiWrite<void>("/system/cancel-next-bell", {
    method: "POST" as const,
  });
}
