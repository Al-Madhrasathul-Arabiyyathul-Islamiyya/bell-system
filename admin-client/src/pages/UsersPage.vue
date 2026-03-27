<script setup lang="ts">
import { computed, ref } from "vue";
import { Icon } from "@iconify/vue";
import { useStorage } from "@vueuse/core";
import AppConfirmDialog from "../components/app/AppConfirmDialog.vue";
import { AppPageHeader, AppStatCard } from "./page-exports";
import UserFormDialog from "../features/users/components/UserFormDialog.vue";
import UsersTable from "../features/users/components/UsersTable.vue";
import {
  useCreateUserMutation,
  useDeleteUserMutation,
  useUpdateUserMutation,
  useUsersQuery,
} from "../features/users/composables/use-users";
import { getUserFacingError } from "../lib/api/errors";
import type { UserRole } from "../lib/api/types";
import type { UserVm, UserWritePayload } from "../lib/api/users";
import { useToastStore } from "../stores/toast";

type RoleFilter = "all" | UserRole;
type UserFormPayload = {
  password?: string;
  role: UserRole;
  username: string;
};

const toastStore = useToastStore();

const currentPage = ref(1);
const roleFilter = useStorage<RoleFilter>(
  "bell-admin-users-role-filter",
  "all",
);
const pageSize = useStorage<number>("bell-admin-users-page-size", 20);

const dialogMode = ref<"create" | "edit">("create");
const dialogOpen = ref(false);
const dialogUser = ref<null | UserVm>(null);
const dialogError = ref<null | string>(null);

const deleteTarget = ref<null | UserVm>(null);

const queryParams = computed(() => ({
  page: currentPage.value,
  role: roleFilter.value === "all" ? undefined : roleFilter.value,
  size: pageSize.value,
  sort: "username",
}));

const usersQuery = useUsersQuery(queryParams);
const createUserMutation = useCreateUserMutation();
const updateUserMutation = useUpdateUserMutation();
const deleteUserMutation = useDeleteUserMutation();

const users = computed(() => usersQuery.data.value?.items ?? []);
const totalUsers = computed(() => readMetaTotal(usersQuery.data.value?.meta));
const totalPages = computed(() => readMetaPages(usersQuery.data.value?.meta));
const adminUsersOnPage = computed(() => {
  return users.value.filter((user) => user.role === "admin").length;
});
const currentMutationPending = computed(() => {
  return (
    createUserMutation.isPending.value || updateUserMutation.isPending.value
  );
});
const pageError = computed(() => {
  if (!usersQuery.error.value) {
    return null;
  }

  return getUserFacingError(usersQuery.error.value).detail;
});

function openCreateDialog() {
  dialogMode.value = "create";
  dialogUser.value = null;
  dialogError.value = null;
  dialogOpen.value = true;
}

function openEditDialog(user: UserVm) {
  dialogMode.value = "edit";
  dialogUser.value = user;
  dialogError.value = null;
  dialogOpen.value = true;
}

function closeDialog() {
  dialogOpen.value = false;
  dialogError.value = null;
}

async function handleSaveUser(payload: UserFormPayload) {
  dialogError.value = null;

  try {
    if (dialogMode.value === "create") {
      await createUserMutation.mutateAsync(toUserWritePayload(payload));
      toastStore.enqueue({
        detail: `${payload.username} was created successfully.`,
        title: "User Created",
        tone: "success",
      });
    } else if (dialogUser.value) {
      await updateUserMutation.mutateAsync({
        id: dialogUser.value.id,
        payload: toUserWritePayload(payload),
      });
      toastStore.enqueue({
        detail: `${payload.username} was updated successfully.`,
        title: "User Updated",
        tone: "success",
      });
    }

    dialogOpen.value = false;
    dialogUser.value = null;
  } catch (error) {
    dialogError.value = getUserFacingError(error).detail;
  }
}

function requestDelete(user: UserVm) {
  deleteTarget.value = user;
}

function cancelDelete() {
  deleteTarget.value = null;
}

async function confirmDelete() {
  if (!deleteTarget.value) {
    return;
  }

  try {
    await deleteUserMutation.mutateAsync(deleteTarget.value.id);
    toastStore.enqueue({
      detail: `${deleteTarget.value.username} was removed successfully.`,
      title: "User Deleted",
      tone: "success",
    });
    deleteTarget.value = null;
  } catch (error) {
    toastStore.enqueue({
      detail: getUserFacingError(error).detail,
      title: "Delete Failed",
      tone: "error",
    });
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value += 1;
  }
}

function previousPage() {
  if (currentPage.value > 1) {
    currentPage.value -= 1;
  }
}

function onRoleFilterChange(value: RoleFilter) {
  roleFilter.value = value;
  currentPage.value = 1;
}

function onPageSizeChange(value: number) {
  pageSize.value = value;
  currentPage.value = 1;
}

function toUserWritePayload(payload: UserFormPayload): UserWritePayload {
  return {
    data: {
      attributes: {
        ...(payload.password ? { password: payload.password } : {}),
        role: payload.role,
        username: payload.username,
      },
      type: "users",
    },
  };
}

function readMetaTotal(meta: Record<string, unknown> | undefined) {
  const value = meta?.total;

  return typeof value === "number" ? value : 0;
}

function readMetaPages(meta: Record<string, unknown> | undefined) {
  const value = meta?.page;

  if (
    value &&
    typeof value === "object" &&
    "pages" in value &&
    typeof value.pages === "number"
  ) {
    return Math.max(value.pages, 1);
  }

  return 1;
}
</script>

<template>
  <section class="page-stack">
    <AppPageHeader
      eyebrow="Administration"
      title="Users"
      description="Manage administrator and operator accounts, filter by role, and keep user credentials aligned with the backend roles."
    />

    <div class="grid gap-4 xl:grid-cols-3">
      <AppStatCard
        title="Total Users"
        icon="solar:users-group-rounded-bold-duotone"
        :value="String(totalUsers)"
        description="All user accounts returned by the backend pagination metadata."
      />
      <AppStatCard
        title="Admins on This Page"
        icon="solar:shield-user-bold-duotone"
        :value="String(adminUsersOnPage)"
        description="Visible administrator accounts in the current result set."
      />
    </div>

    <article class="card border border-base-300 bg-base-100 shadow-sm">
      <div class="card-body gap-5">
        <div
          class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between"
        >
          <div class="grid gap-4 md:grid-cols-2">
            <fieldset class="fieldset">
              <legend class="fieldset-legend">Role</legend>
              <select
                class="select w-full"
                :value="roleFilter"
                @change="
                  onRoleFilterChange(
                    ($event.target as HTMLSelectElement).value as RoleFilter,
                  )
                "
              >
                <option value="all">All Roles</option>
                <option value="admin">Admin</option>
                <option value="morning_user">Morning User</option>
                <option value="afternoon_user">Afternoon User</option>
              </select>
            </fieldset>

            <fieldset class="fieldset">
              <legend class="fieldset-legend">Page Size</legend>
              <select
                class="select w-full"
                :value="String(pageSize)"
                @change="
                  onPageSizeChange(
                    Number(($event.target as HTMLSelectElement).value),
                  )
                "
              >
                <option value="10">10 per page</option>
                <option value="20">20 per page</option>
                <option value="50">50 per page</option>
              </select>
            </fieldset>
          </div>

          <button
            class="btn btn-primary"
            type="button"
            @click="openCreateDialog"
          >
            <Icon icon="solar:user-plus-rounded-bold-duotone" class="text-lg" />
            Create User
          </button>
        </div>

        <div v-if="pageError" class="alert alert-error text-sm" role="alert">
          {{ pageError }}
        </div>

        <UsersTable
          :users="users"
          :loading="usersQuery.isLoading.value || usersQuery.isFetching.value"
          :deleting-id="deleteUserMutation.variables.value ?? null"
          @edit="openEditDialog"
          @delete="requestDelete"
        />

        <div
          class="flex flex-col gap-3 border-t border-base-300 pt-4 sm:flex-row sm:items-center sm:justify-between"
        >
          <p class="text-sm text-base-content/65">
            Showing {{ users.length }} of {{ totalUsers }} users.
          </p>
          <div class="join">
            <button
              class="btn join-item"
              type="button"
              :disabled="currentPage <= 1 || usersQuery.isFetching.value"
              @click="previousPage"
            >
              Previous
            </button>
            <button class="btn join-item pointer-events-none" type="button">
              {{ currentPage }}
            </button>
            <button
              class="btn join-item"
              type="button"
              :disabled="
                currentPage >= totalPages || usersQuery.isFetching.value
              "
              @click="nextPage"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </article>

    <UserFormDialog
      :open="dialogOpen"
      :mode="dialogMode"
      :user="dialogUser"
      :pending="currentMutationPending"
      :error="dialogError"
      @close="closeDialog"
      @save="handleSaveUser"
    />

    <AppConfirmDialog
      :open="Boolean(deleteTarget)"
      title="Delete User"
      :description="
        deleteTarget
          ? `Remove ${deleteTarget.username} from the system. This action cannot be undone.`
          : ''
      "
      confirm-label="Delete User"
      tone="error"
      :pending="deleteUserMutation.isPending.value"
      @close="cancelDelete"
      @confirm="confirmDelete"
    />
  </section>
</template>
