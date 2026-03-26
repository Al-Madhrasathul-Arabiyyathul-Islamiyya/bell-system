<script setup lang="ts">
import { computed, reactive, watch } from "vue";
import { useRegleSchema } from "@regle/schemas";
import { z } from "zod";
import type { UserRole } from "../../../lib/api/types";
import type { UserVm } from "../../../lib/api/users";

type UserFormPayload = {
  password?: string;
  role: UserRole;
  username: string;
};

const props = withDefaults(
  defineProps<{
    error?: null | string;
    mode: "create" | "edit";
    open: boolean;
    pending?: boolean;
    user?: null | UserVm;
  }>(),
  {
    error: null,
    pending: false,
    user: null,
  },
);

const emit = defineEmits<{
  close: [];
  save: [payload: UserFormPayload];
}>();

const form = reactive({
  password: "",
  role: "morning_user" as UserRole,
  username: "",
});

const schema = computed(() => {
  if (props.mode === "create") {
    return z.object({
      password: z
        .string()
        .trim()
        .min(8, "Password must be at least 8 characters."),
      role: z.enum(["admin", "morning_user", "afternoon_user"]),
      username: z.string().trim().min(1, "Username is required."),
    });
  }

  return z.object({
    password: z
      .string()
      .trim()
      .refine(
        (value) => value.length === 0 || value.length >= 8,
        "Password must be at least 8 characters.",
      ),
    role: z.enum(["admin", "morning_user", "afternoon_user"]),
    username: z.string().trim().min(1, "Username is required."),
  });
});

const { r$ } = useRegleSchema(form, schema);

const dialogTitle = computed(() => {
  return props.mode === "create" ? "Create User" : "Edit User";
});

const submitLabel = computed(() => {
  if (props.pending) {
    return props.mode === "create" ? "Creating..." : "Saving...";
  }

  return props.mode === "create" ? "Create User" : "Save Changes";
});

watch(
  () => [props.mode, props.open, props.user?.id] as const,
  () => {
    if (!props.open) {
      return;
    }

    form.username = props.user?.username ?? "";
    form.role = props.user?.role ?? "morning_user";
    form.password = "";
    r$.$reset();
  },
  { immediate: true },
);

async function handleSubmit() {
  const result = await r$.$validate();

  if (!result.valid) {
    return;
  }

  const password = form.password.trim();

  emit("save", {
    ...(password ? { password } : {}),
    role: form.role,
    username: form.username.trim(),
  });
}
</script>

<template>
  <div v-if="open" class="modal modal-open">
    <div class="modal-box max-w-lg">
      <button
        class="btn btn-circle btn-ghost btn-sm absolute top-4 right-4"
        type="button"
        :disabled="pending"
        aria-label="Close user form"
        @click="emit('close')"
      >
        ×
      </button>

      <h2 class="font-display pr-8 text-2xl font-semibold text-base-content">
        {{ dialogTitle }}
      </h2>
      <p class="mt-2 text-sm leading-7 text-base-content/70">
        {{
          mode === "create"
            ? "Create a new account and assign its role."
            : "Update the username, role, or password for this account."
        }}
      </p>

      <div v-if="error" class="alert alert-error mt-5 text-sm" role="alert">
        {{ error }}
      </div>

      <form class="mt-6 space-y-5" @submit.prevent="handleSubmit">
        <fieldset class="fieldset">
          <legend class="fieldset-legend">Username</legend>
          <input
            v-model="r$.$value.username"
            class="input w-full"
            :disabled="pending"
            type="text"
            placeholder="username"
          />
          <p v-if="r$.username.$error" class="fieldset-label text-error">
            {{ r$.username.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset">
          <legend class="fieldset-legend">Role</legend>
          <select
            v-model="r$.$value.role"
            class="select w-full"
            :disabled="pending"
          >
            <option value="admin">Admin</option>
            <option value="morning_user">Morning User</option>
            <option value="afternoon_user">Afternoon User</option>
          </select>
          <p v-if="r$.role.$error" class="fieldset-label text-error">
            {{ r$.role.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset">
          <legend class="fieldset-legend">
            {{ mode === "create" ? "Password" : "New Password" }}
          </legend>
          <input
            v-model="r$.$value.password"
            class="input w-full"
            :disabled="pending"
            type="password"
            :placeholder="
              mode === 'create'
                ? 'Minimum 8 characters'
                : 'Leave blank to keep the current password'
            "
          />
          <p class="fieldset-label">
            {{
              mode === "create"
                ? "At least 8 characters."
                : "Leave blank to keep the existing password."
            }}
          </p>
          <p v-if="r$.password.$error" class="fieldset-label text-error">
            {{ r$.password.$errors[0] }}
          </p>
        </fieldset>

        <div class="modal-action mt-8">
          <button
            class="btn btn-ghost"
            type="button"
            :disabled="pending"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button class="btn btn-primary" type="submit" :disabled="pending">
            {{ submitLabel }}
          </button>
        </div>
      </form>
    </div>
    <div class="modal-backdrop" @click="emit('close')" />
  </div>
</template>
