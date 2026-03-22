const DEFAULT_API_BASE_URL = "http://localhost:8080/api/v1";
const DEFAULT_WS_BASE_URL = "ws://localhost:8080/ws";
const DEFAULT_APP_NAME = "Bell System Admin";

export const appEnv = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || DEFAULT_API_BASE_URL,
  appName: import.meta.env.VITE_APP_NAME || DEFAULT_APP_NAME,
  wsBaseUrl: import.meta.env.VITE_WS_BASE_URL || DEFAULT_WS_BASE_URL,
} as const;
