import { defineConfig } from "@apps-in-toss/web-framework/config";

export default defineConfig({
  appName: process.env.AIT_APP_NAME ?? "habitcashback", // 콘솔 AppName(=App ID)와 동일하게 설정
  brand: {
    displayName: "습관환급",
    primaryColor: "#3182F6",
    icon: process.env.AIT_ICON_URL ?? "", // 콘솔에 업로드한 아이콘 URL (선택)
    bridgeColorMode: "basic",
  },
  enableTossPay: true, // 토스페이 결제 활성화
  navigationBar: {
    backgroundColor: "#FFFFFF",
    titleColor: "#191F28",
    type: "default",
  },
  web: {
    host: "localhost",
    port: 5173,
    commands: {
      dev: "vite",
      build: "vite build",
    },
  },
  permissions: [
    { name: "camera", access: "access" },
    { name: "photos", access: "read" }
  ],
  outdir: "dist",
  webViewProps: {
    type: "partner",
    bounces: false,
    pullToRefreshEnabled: false,
    overScrollMode: "never",
  },
});
