import { toValue, type MaybeRefOrGetter } from "vue";
import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
import {
  deleteAudio,
  getAudioChecksums,
  getAudioContent,
  getAudioFile,
  listAudioFiles,
  updateAudio,
  uploadAudio,
  type AudioListParams,
  type AudioUpdatePayload,
} from "../../../lib/api/audio";
import { queryKeys } from "../../../lib/api/query-keys";

export function useAudioFilesQuery(
  params: MaybeRefOrGetter<AudioListParams> = {},
) {
  return useQuery({
    queryFn: () => listAudioFiles(toValue(params)),
    queryKey: queryKeys.audio(toValue(params)),
  });
}

export function useAudioFileQuery(id: string) {
  return useQuery({
    enabled: Boolean(id),
    queryFn: () => getAudioFile(id),
    queryKey: [...queryKeys.audio(), "detail", id] as const,
  });
}

export function useAudioChecksumsQuery() {
  return useQuery({
    queryFn: () => getAudioChecksums(),
    queryKey: queryKeys.audioChecksums(),
  });
}

export function useUploadAudioMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: FormData) => uploadAudio(payload),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ["audio"] });
      queryClient.invalidateQueries({ queryKey: queryKeys.audioChecksums() });
    },
  });
}

export function useUpdateAudioMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      payload,
    }: {
      id: string;
      payload: AudioUpdatePayload;
    }) => updateAudio(id, payload),
    onSuccess(_, variables) {
      queryClient.invalidateQueries({ queryKey: ["audio"] });
      queryClient.invalidateQueries({
        queryKey: [...queryKeys.audio(), "detail", variables.id] as const,
      });
    },
  });
}

export function useDeleteAudioMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteAudio(id),
    onSuccess() {
      queryClient.invalidateQueries({ queryKey: ["audio"] });
      queryClient.invalidateQueries({ queryKey: queryKeys.audioChecksums() });
    },
  });
}

export function useAudioContentQuery(id: string) {
  return useQuery({
    enabled: Boolean(id),
    queryFn: () => getAudioContent(id),
    queryKey: [...queryKeys.audio(), "content", id] as const,
  });
}
