import React from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import { RequireAuth } from "./Auth";
import LoginPage from "../pages/LoginPage";
import HomePage from "../pages/HomePage";
import ChallengeDetailPage from "../pages/ChallengeDetailPage";
import ProofPage from "../pages/ProofPage";
import HistoryPage from "../pages/HistoryPage";
import HelpPage from "../pages/HelpPage";
import TermsPage from "../pages/TermsPage";
import PrivacyPage from "../pages/PrivacyPage";
import SupportPage from "../pages/SupportPage";

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/help" element={<HelpPage />} />
      <Route path="/terms" element={<TermsPage />} />
      <Route path="/privacy" element={<PrivacyPage />} />
      <Route path="/support" element={<SupportPage />} />

      <Route
        path="/"
        element={
          <RequireAuth>
            <HomePage />
          </RequireAuth>
        }
      />
      <Route
        path="/challenge/:id"
        element={
          <RequireAuth>
            <ChallengeDetailPage />
          </RequireAuth>
        }
      />
      <Route
        path="/proof/:id"
        element={
          <RequireAuth>
            <ProofPage />
          </RequireAuth>
        }
      />
      <Route
        path="/history"
        element={
          <RequireAuth>
            <HistoryPage />
          </RequireAuth>
        }
      />

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
