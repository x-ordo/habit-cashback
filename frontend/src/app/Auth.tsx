import React from "react";
import { Navigate, useLocation } from "react-router-dom";
import { getAccessToken } from "../lib/storage";

export function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = getAccessToken();
  const loc = useLocation();

  if (!token) {
    return <Navigate to="/login" replace state={{ from: loc.pathname }} />;
  }
  return <>{children}</>;
}
