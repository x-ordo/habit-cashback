import React, { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Button, Text } from "@toss-design-system/mobile";
import { appLogin } from "@apps-in-toss/web-framework";
import LegalFooter from "../components/LegalFooter";
import { apiPost } from "../lib/api";
import { setAccessToken } from "../lib/storage";

type ExchangeResponse = {
  accessToken?: string;
  sessionToken?: string;
  mode?: "stub" | "toss";
};

export default function LoginPage() {
  const loc = useLocation();
  const nav = useNavigate();
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const login = async () => {
    setLoading(true);
    setErr(null);

    try {
      // 1) Apps in Toss: get authorizationCode via SDK, exchange on our server (mTLS required on server).
      try {
        const { authorizationCode, referrer } = await appLogin();

        const resp = await apiPost<ExchangeResponse>("/v1/auth/toss/exchange", {
          authorizationCode,
          referrer,
        });

        const token = resp.sessionToken ?? resp.accessToken;
        if (!token) throw new Error("token missing");
        setAccessToken(token);
        nav("/challenges");
        return;
      } catch (e) {
        // 2) Local browser / non-Toss env: fallback to stub login for development & review.
        const resp = await apiPost<ExchangeResponse>("/v1/auth/exchange", {});
        const token = resp.sessionToken ?? resp.accessToken;
        if (!token) throw new Error("token missing");
        setAccessToken(token);
        nav("/challenges");
        return;
      }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : "로그인에 실패했습니다.";
      setErr(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: 16, maxWidth: 520, margin: "0 auto" }}>
      <Text typography="t2" fontWeight="bold">
        로그인
      </Text>
      
{new URLSearchParams(loc.search).get("reason") === "unlinked" && (
  <>
    <div style={{ height: 12 }} />
    <Text typography="t6" color="red500">
      토스 연결이 해제되어 다시 로그인이 필요합니다.
    </Text>
  </>
)}
<div style={{ height: 8 }} />
      <Text typography="t6" color="grey600">
        토스앱에서는 토스 로그인을 통해 인가 코드를 받고, 서버에서 토큰으로 교환합니다.
      </Text>

      <div style={{ height: 16 }} />

      {err && (
        <>
          <Text typography="t6" color="red500">
            {err}
          </Text>
          <div style={{ height: 12 }} />
        </>
      )}

      <Button size="large" style={{ width: "100%" }} onClick={login} disabled={loading}>
        {loading ? "로그인 중..." : "토스 로그인"}
      </Button>

      <div style={{ height: 12 }} />
      <Text typography="t7" color="grey600">
        * 앱인토스 로그인(appLogin) 인가 코드 유효시간은 10분입니다.
        * 로컬 브라우저에서는 데모 로그인으로 자동 대체됩니다.
      </Text>

      <div style={{ height: 12 }} />
      <LegalFooter />
    </div>
  );
}
