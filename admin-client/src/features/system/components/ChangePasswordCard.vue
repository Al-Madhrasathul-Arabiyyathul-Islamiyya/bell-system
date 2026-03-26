<script setup lang="ts">
import { computed, reactive } from "vue";
import { useRegleSchema } from "@regle/schemas";
import { z } from "zod";

type ChangePasswordPayload = {
  newPassword: string;
  oldPassword: string;
};

const props = withDefaults(
  defineProps<{
    error?: null | string;
    pending?: boolean;
  }>(),
  {
    error: null,
    pending: false,
  },
);

const emit = defineEmits<{
  save: [payload: ChangePasswordPayload];
}>();

const form = reactive({
  confirmPassword: "",
  newPassword: "",
  oldPassword: "",
});

const schema = z
  .object({
    confirmPassword: z
      .string()
      .trim()
      .min(1, "Please confirm the new password."),
    newPassword: z
      .string()
      .trim()
      .min(8, "New password must be at least 8 characters."),
    oldPassword: z.string().trim().min(1, "Current password is required."),
  })
  .refine((value) => value.newPassword !== value.oldPassword, {
    message: "New password must be different from the current password.",
    path: ["newPassword"],
  })
  .refine((value) => value.newPassword === value.confirmPassword, {
    message: "New password and confirmation must match.",
    path: ["confirmPassword"],
  });

const { r$ } = useRegleSchema(form, schema);

const submitLabel = computed(() => {
  return props.pending ? "Updating..." : "Change Password";
});

async function handleSubmit() {
  const result = await r$.$validate();

  if (!result.valid) {
    return;
  }

  emit("save", {
    newPassword: form.newPassword.trim(),
    oldPassword: form.oldPassword.trim(),
  });
}

function reset() {
  form.oldPassword = "";
  form.newPassword = "";
  form.confirmPassword = "";
  r$.$reset();
}

defineExpose({
  reset,
});
</script>

<template>
  <article class="card border border-base-300 bg-base-100 shadow-sm">
    <div class="card-body gap-5">
      <div class="space-y-2">
        <p
          class="text-xs font-semibold uppercase tracking-[0.24em] text-primary"
        >
          Account Security
        </p>
        <h2 class="font-display text-3xl font-semibold text-base-content">
          Change Password
        </h2>
        <p class="text-sm leading-7 text-base-content/70">
          Update the password for the current account without leaving the admin
          client.
        </p>
      </div>

      <div v-if="error" class="alert alert-error text-sm" role="alert">
        {{ error }}
      </div>

      <form class="grid gap-4 md:grid-cols-2" @submit.prevent="handleSubmit">
        <fieldset class="fieldset md:col-span-2">
          <legend class="fieldset-legend">Current password</legend>
          <input
            v-model="r$.$value.oldPassword"
            class="input w-full"
            :disabled="pending"
            type="password"
            placeholder="Current password"
          />
          <p v-if="r$.oldPassword.$error" class="fieldset-label text-error">
            {{ r$.oldPassword.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset">
          <legend class="fieldset-legend">New password</legend>
          <input
            v-model="r$.$value.newPassword"
            class="input w-full"
            :disabled="pending"
            type="password"
            placeholder="Minimum 8 characters"
          />
          <p v-if="r$.newPassword.$error" class="fieldset-label text-error">
            {{ r$.newPassword.$errors[0] }}
          </p>
        </fieldset>

        <fieldset class="fieldset">
          <legend class="fieldset-legend">Confirm new password</legend>
          <input
            v-model="r$.$value.confirmPassword"
            class="input w-full"
            :disabled="pending"
            type="password"
            placeholder="Repeat new password"
          />
          <p v-if="r$.confirmPassword.$error" class="fieldset-label text-error">
            {{ r$.confirmPassword.$errors[0] }}
          </p>
        </fieldset>

        <div class="md:col-span-2 flex justify-end">
          <button class="btn btn-primary" type="submit" :disabled="pending">
            {{ submitLabel }}
          </button>
        </div>
      </form>
    </div>
  </article>
</template>
