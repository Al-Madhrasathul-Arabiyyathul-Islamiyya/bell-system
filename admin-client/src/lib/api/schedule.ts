import { adaptJsonApiCollection } from "../jsonapi/adapters";
import { getIncludedResource, getRelationshipData } from "../jsonapi/documents";
import { apiJson, apiJsonApiCollection, apiJsonApiWrite } from "./client";
import type { JsonApiCollectionDocument, JsonApiDocument } from "./types";

type ScheduleItemAttributes = {
  createdAt: string;
  days: number[];
  name: string;
  time: string;
  updatedAt: string;
};

type ScheduleWriteAttributes = {
  days: number[];
  name: string;
  time: string;
};

type SessionAttributes = {
  endTime: string;
  name: string;
  startTime: string;
};

type AudioAttributes = {
  checksum: string;
  createdAt: string;
  fileType: "anthem" | "bell" | "other" | "school_song";
  name: string;
  updatedAt: string;
};

export type ScheduleItemVm = {
  createdAt: string;
  days: number[];
  id: string;
  name: string;
  sessionId: string | null;
  sessionName: string | null;
  soundId: string | null;
  soundName: string | null;
  time: string;
  updatedAt: string;
};

export type ScheduleListParams = {
  day?: number;
  include?: string;
  sessionId?: string;
};

export type CurrentScheduleResponse = {
  items: Array<{
    id: string;
    name: string;
    sound: {
      checksum: string;
      fileType: "anthem" | "bell" | "other" | "school_song";
      id: string;
      name: string;
    } | null;
    status?: string;
    time: string;
  }>;
  session: {
    endTime?: string;
    id?: string;
    name?: string;
    startTime?: string;
  } | null;
};

export type ScheduleWritePayload = {
  data: {
    attributes: ScheduleWriteAttributes;
    relationships: {
      session: {
        data: null | {
          id: string;
          type: "sessions";
        };
      };
      sound: {
        data: {
          id: string;
          type: "audio-files";
        };
      };
    };
    type: "schedule-items";
  };
};

export async function listScheduleItems(params: ScheduleListParams = {}) {
  const include = params.include ?? "session,sound";
  const document = await apiJsonApiCollection<ScheduleItemAttributes>(
    "/schedule",
    {
      query: {
        "filter[day]": params.day,
        "filter[sessionId]": params.sessionId,
        include,
      },
    },
  );

  return adaptJsonApiCollection(document, mapScheduleItem);
}

export async function getCurrentSchedule() {
  return apiJson<CurrentScheduleResponse>("/schedule/current");
}

export async function createScheduleItem(payload: ScheduleWritePayload) {
  return apiJsonApiWrite<JsonApiDocument<ScheduleItemAttributes>>("/schedule", {
    body: payload,
    method: "POST",
  });
}

export async function updateScheduleItem(
  id: string,
  payload: ScheduleWritePayload,
) {
  return apiJsonApiWrite<JsonApiDocument<ScheduleItemAttributes>>(
    `/schedule/${id}`,
    {
      body: payload,
      method: "PUT",
    },
  );
}

export async function deleteScheduleItem(id: string) {
  return apiJsonApiWrite<void>(`/schedule/${id}`, {
    method: "DELETE",
  });
}

function mapScheduleItem(
  resource: JsonApiCollectionDocument<ScheduleItemAttributes>["data"][number],
  document: JsonApiCollectionDocument<ScheduleItemAttributes>,
): ScheduleItemVm {
  const sessionRelationship = getRelationshipData(
    resource.relationships,
    "session",
  );
  const soundRelationship = getRelationshipData(
    resource.relationships,
    "sound",
  );
  const sessionResource =
    !Array.isArray(sessionRelationship) && sessionRelationship
      ? getIncludedResource(document, sessionRelationship)
      : null;
  const soundResource =
    !Array.isArray(soundRelationship) && soundRelationship
      ? getIncludedResource(document, soundRelationship)
      : null;

  return {
    createdAt: resource.attributes.createdAt,
    days: resource.attributes.days,
    id: resource.id,
    name: resource.attributes.name,
    sessionId:
      !Array.isArray(sessionRelationship) && sessionRelationship
        ? sessionRelationship.id
        : null,
    sessionName:
      (sessionResource?.attributes as SessionAttributes | undefined)?.name ??
      null,
    soundId:
      !Array.isArray(soundRelationship) && soundRelationship
        ? soundRelationship.id
        : null,
    soundName:
      (soundResource?.attributes as AudioAttributes | undefined)?.name ?? null,
    time: resource.attributes.time,
    updatedAt: resource.attributes.updatedAt,
  };
}
