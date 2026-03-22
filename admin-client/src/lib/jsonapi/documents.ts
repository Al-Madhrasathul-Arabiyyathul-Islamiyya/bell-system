import type {
  JsonApiCollectionDocument,
  JsonApiDocument,
  JsonApiRelationship,
  JsonApiResource,
} from "../api/types";

function includeKey(resource: { id: string; type: string }) {
  return `${resource.type}:${resource.id}`;
}

export function getIncludedMap(document: { included?: JsonApiResource[] }) {
  return new Map(
    (document.included ?? []).map((resource) => [
      includeKey(resource),
      resource,
    ]),
  );
}

export function getRelationshipData(
  relationships: Record<string, JsonApiRelationship> | undefined,
  name: string,
) {
  return relationships?.[name]?.data ?? null;
}

export function getIncludedResource(
  document: JsonApiDocument | JsonApiCollectionDocument,
  relationship: null | {
    id: string;
    type: string;
  },
) {
  if (!relationship) {
    return null;
  }

  return getIncludedMap(document).get(includeKey(relationship)) ?? null;
}
