import { adaptJsonApiCollection } from "../jsonapi/adapters";
import {
  apiBinary,
  apiJson,
  apiJsonApiCollection,
  apiJsonApiDocument,
  apiJsonApiWrite,
  apiMultipart,
} from "./client";
import type { JsonApiDocument } from "./types";

export type AudioFileType = "anthem" | "bell" | "other" | "school_song";

export type AudioAttributes = {
  checksum: string;
  createdAt: string;
  fileType: AudioFileType;
  name: string;
  updatedAt: string;
};

export type AudioFileVm = {
  checksum: string;
  contentUrl: string;
  createdAt: string;
  fileType: AudioFileType;
  id: string;
  name: string;
  updatedAt: string;
};

export type AudioListParams = {
  fields?: string;
  fileType?: AudioFileType;
  page?: number;
  size?: number;
  sort?: string;
};

export type AudioUpdatePayload = {
  data: {
    attributes: {
      fileType: AudioFileType;
      name: string;
    };
    type: "audio-files";
  };
};

export async function listAudioFiles(params: AudioListParams = {}) {
  const document = await apiJsonApiCollection<AudioAttributes>("/audio", {
    query: {
      "fields[audio-files]": params.fields,
      "filter[fileType]": params.fileType,
      "page[number]": params.page,
      "page[size]": params.size,
      sort: params.sort,
    },
  });

  return adaptJsonApiCollection(document, (resource) => ({
    checksum: resource.attributes.checksum,
    contentUrl: `/api/v1/audio/${resource.id}/content`,
    createdAt: resource.attributes.createdAt,
    fileType: resource.attributes.fileType,
    id: resource.id,
    name: resource.attributes.name,
    updatedAt: resource.attributes.updatedAt,
  }));
}

export async function getAudioFile(id: string) {
  const document = await apiJsonApiDocument<AudioAttributes>(`/audio/${id}`);

  return {
    checksum: document.data.attributes.checksum,
    contentUrl: `/api/v1/audio/${id}/content`,
    createdAt: document.data.attributes.createdAt,
    fileType: document.data.attributes.fileType,
    id: document.data.id,
    name: document.data.attributes.name,
    updatedAt: document.data.attributes.updatedAt,
  } satisfies AudioFileVm;
}

export async function getAudioChecksums() {
  return apiJson<Record<string, string>>("/audio/checksums");
}

export async function uploadAudio(payload: FormData) {
  return apiMultipart<void>("/audio", {
    body: payload,
    method: "POST",
  });
}

export async function updateAudio(id: string, payload: AudioUpdatePayload) {
  return apiJsonApiWrite<JsonApiDocument<AudioAttributes>>(`/audio/${id}`, {
    body: payload,
    method: "PUT",
  });
}

export async function deleteAudio(id: string) {
  return apiJsonApiWrite<void>(`/audio/${id}`, {
    method: "DELETE",
  });
}

export async function getAudioContent(id: string) {
  return apiBinary(`/audio/${id}/content`);
}
