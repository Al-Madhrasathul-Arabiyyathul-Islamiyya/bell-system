import { createRequire } from "node:module";
import js from "@eslint/js";
import globals from "globals";
import pluginVue from "eslint-plugin-vue";
import tseslint from "typescript-eslint";
import eslintConfigPrettier from "eslint-config-prettier";

const require = createRequire(import.meta.url);
const vueEslintParser = require("vue-eslint-parser");

export default tseslint.config(
  {
    ignores: ["dist/**", "node_modules/**", "*.d.ts"],
  },
  js.configs.recommended,
  ...pluginVue.configs["flat/recommended"],
  ...tseslint.configs.recommended,
  {
    files: ["**/*.{ts,tsx,mts,vue}"],
    languageOptions: {
      parser: vueEslintParser,
      ecmaVersion: "latest",
      sourceType: "module",
      globals: {
        ...globals.browser,
        ...globals.node,
      },
      parserOptions: {
        parser: tseslint.parser,
        extraFileExtensions: [".vue"],
      },
    },
    rules: {
      "vue/multi-word-component-names": "off",
    },
  },
  eslintConfigPrettier,
);
