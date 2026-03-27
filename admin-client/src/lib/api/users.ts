import { adaptJsonApiCollection } from "../jsonapi/adapters";
import {
  apiJsonApiCollection,
  apiJsonApiDocument,
  apiJsonApiWrite,
} from "./client";
import type { JsonApiDocument, UserRole } from "./types";

export type UserAttributes = {
  createdAt: string;
  role: UserRole;
  username: string;
};

export type UserVm = {
  createdAt: string;
  id: string;
  role: UserRole;
  username: string;
};

export type UsersQueryParams = {
  page?: number;
  role?: UserRole;
  size?: number;
  sort?: string;
};

export type UserWritePayload = {
  data: {
    attributes: {
      password?: string;
      role: UserRole;
      username: string;
    };
    type: "users";
  };
};

export async function listUsers(params: UsersQueryParams = {}) {
  const document = await apiJsonApiCollection<UserAttributes>("/users", {
    query: {
      "filter[role]": params.role,
      "page[number]": params.page,
      "page[size]": params.size,
      sort: params.sort,
    },
  });

  return adaptJsonApiCollection(document, (resource) => ({
    createdAt: resource.attributes.createdAt,
    id: resource.id,
    role: resource.attributes.role,
    username: resource.attributes.username,
  }));
}

export async function getUser(id: string) {
  const document = await apiJsonApiDocument<UserAttributes>(`/users/${id}`);

  return mapUser(document);
}

export async function createUser(payload: UserWritePayload) {
  return apiJsonApiWrite<JsonApiDocument<UserAttributes>>("/users", {
    body: payload,
    method: "POST",
  });
}

export async function updateUser(id: string, payload: UserWritePayload) {
  return apiJsonApiWrite<JsonApiDocument<UserAttributes>>(`/users/${id}`, {
    body: payload,
    method: "PUT",
  });
}

export async function deleteUser(id: string) {
  return apiJsonApiWrite<void>(`/users/${id}`, {
    method: "DELETE",
  });
}

function mapUser(document: JsonApiDocument<UserAttributes>): UserVm {
  return {
    createdAt: document.data.attributes.createdAt,
    id: document.data.id,
    role: document.data.attributes.role,
    username: document.data.attributes.username,
  };
}
