import type {
  JsonApiCollectionDocument,
  JsonApiDocument,
  JsonApiListResult,
  JsonApiResource,
} from "../api/types";

export function adaptJsonApiResource<TAttributes, TResult>(
  document: JsonApiDocument<TAttributes>,
  mapper: (
    resource: JsonApiResource<TAttributes>,
    document: JsonApiDocument<TAttributes>,
  ) => TResult,
) {
  return mapper(document.data, document);
}

export function adaptJsonApiCollection<TAttributes, TResult>(
  document: JsonApiCollectionDocument<TAttributes>,
  mapper: (
    resource: JsonApiResource<TAttributes>,
    document: JsonApiCollectionDocument<TAttributes>,
  ) => TResult,
): JsonApiListResult<TResult> {
  return {
    items: document.data.map((resource) => mapper(resource, document)),
    links: document.links,
    meta: document.meta,
  };
}
