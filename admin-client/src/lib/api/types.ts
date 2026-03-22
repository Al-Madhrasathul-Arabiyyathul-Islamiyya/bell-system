export type UserRole = "admin" | "morning_user" | "afternoon_user";

export type AuthUser = {
  id: string | null;
  role: UserRole | null;
  username: string | null;
};

export type AppError = {
  code?: string;
  detail: string;
  source?: {
    parameter?: string;
    pointer?: string;
  };
  status?: number;
  title: string;
};

export type JsonApiResource<TAttributes = Record<string, unknown>> = {
  attributes: TAttributes;
  id: string;
  relationships?: Record<string, JsonApiRelationship>;
  type: string;
};

export type JsonApiRelationship = {
  data:
    | null
    | {
        id: string;
        type: string;
      }
    | Array<{
        id: string;
        type: string;
      }>;
};

export type JsonApiDocument<TAttributes = Record<string, unknown>> = {
  data: JsonApiResource<TAttributes>;
  included?: JsonApiResource[];
  links?: Record<string, unknown>;
  meta?: Record<string, unknown>;
};

export type JsonApiCollectionDocument<TAttributes = Record<string, unknown>> = {
  data: JsonApiResource<TAttributes>[];
  included?: JsonApiResource[];
  links?: Record<string, unknown>;
  meta?: Record<string, unknown>;
};

export type JsonApiErrorDocument = {
  errors: AppError[];
};

export type JsonApiListResult<TItem> = {
  items: TItem[];
  links?: Record<string, unknown>;
  meta?: Record<string, unknown>;
};
