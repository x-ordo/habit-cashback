export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL?.toString().trim() || "";

export const APP_DISPLAY_NAME =
  import.meta.env.VITE_APP_DISPLAY_NAME?.toString().trim() || "습관환급";

export const SUPPORT_EMAIL =
  import.meta.env.VITE_SUPPORT_EMAIL?.toString().trim() || "support@habitrefund.example";

export const SUPPORT_HOURS =
  import.meta.env.VITE_SUPPORT_HOURS?.toString().trim() || "평일 10:00~18:00 (KST)";

export const COMPANY_NAME =
  import.meta.env.VITE_COMPANY_NAME?.toString().trim() || "운영사";

export const PRIVACY_EFFECTIVE_DATE =
  import.meta.env.VITE_PRIVACY_EFFECTIVE_DATE?.toString().trim() || "2025-12-21";
