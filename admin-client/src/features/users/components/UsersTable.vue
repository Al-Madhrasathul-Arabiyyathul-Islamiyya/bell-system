<script setup lang="ts">
import { computed } from "vue";
import { Icon } from "@iconify/vue";
import type { UserVm } from "../../../lib/api/users";
import UserRoleBadge from "./UserRoleBadge.vue";

const props = withDefaults(
  defineProps<{
    deletingId?: null | string;
    loading?: boolean;
    users: UserVm[];
  }>(),
  {
    deletingId: null,
    loading: false,
  },
);

const emit = defineEmits<{
  delete: [user: UserVm];
  edit: [user: UserVm];
}>();

const empty = computed(() => !props.loading && props.users.length === 0);

function formatCreatedAt(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
</script>

<template>
  <div class="overflow-x-auto rounded-box border border-base-300 bg-base-100">
    <table class="table table-zebra">
      <thead>
        <tr>
          <th>Username</th>
          <th>Role</th>
          <th>Created</th>
          <th class="w-32 text-right">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="4">
            <div class="flex items-center justify-center py-10">
              <span class="loading loading-spinner loading-md text-primary" />
            </div>
          </td>
        </tr>

        <tr v-else-if="empty">
          <td colspan="4">
            <div class="flex flex-col items-center gap-3 py-10 text-center">
              <Icon
                icon="solar:user-cross-rounded-bold-duotone"
                class="text-4xl text-base-content/35"
              />
              <div class="space-y-1">
                <p class="text-base font-semibold text-base-content">
                  No users found
                </p>
                <p class="text-sm text-base-content/65">
                  Adjust the role filter or create a new user.
                </p>
              </div>
            </div>
          </td>
        </tr>

        <tr v-for="user in users" :key="user.id">
          <td class="font-medium text-base-content">{{ user.username }}</td>
          <td>
            <UserRoleBadge :role="user.role" />
          </td>
          <td class="text-sm text-base-content/70">
            {{ formatCreatedAt(user.createdAt) }}
          </td>
          <td>
            <div class="flex justify-end gap-2">
              <button
                class="btn btn-sm btn-outline"
                type="button"
                @click="emit('edit', user)"
              >
                Edit
              </button>
              <button
                class="btn btn-sm btn-error btn-soft"
                type="button"
                :disabled="deletingId === user.id"
                @click="emit('delete', user)"
              >
                {{ deletingId === user.id ? "Deleting..." : "Delete" }}
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
