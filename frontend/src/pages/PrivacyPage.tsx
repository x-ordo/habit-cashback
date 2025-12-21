import React from "react";
import { Text } from "@toss-design-system/mobile";
import TopBar from "../components/TopBar";
import LegalFooter from "../components/LegalFooter";
import { APP_DISPLAY_NAME, PRIVACY_EFFECTIVE_DATE } from "../lib/env";

export default function PrivacyPage() {
  return (
    <div style={{ paddingBottom: 24 }}>
      <TopBar title="개인정보처리방침" />
      <div style={{ padding: 16 }}>
        <Text typography="t8" color="grey500">시행일: {PRIVACY_EFFECTIVE_DATE}</Text>

        <div style={{ height: 16 }} />
        <Text typography="t6" fontWeight="bold">1. 수집 항목(최소화)</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          {APP_DISPLAY_NAME}은(는) 서비스 제공에 필요한 최소 정보만 처리합니다. (예: 토큰 기반 식별자, 접속 로그)
          사진 인증 기능을 사용하는 경우, 인증 사진은 부정행위 방지 및 결과 검증 목적에 한해 처리될 수 있습니다.
        </Text>

        <div style={{ height: 16 }} />
        <Text typography="t6" fontWeight="bold">2. 이용 목적</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          (1) 로그인/세션 유지 (2) 챌린지 참여 및 인증 처리 (3) 결과 안내 및 리워드 제공 (4) 부정 사용 방지 (5) CS 대응 및 서비스 개선
        </Text>

        <div style={{ height: 16 }} />
        <Text typography="t6" fontWeight="bold">3. 보관 및 삭제</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          보관 기간은 법령 또는 운영 정책에 따르며, 목적 달성 시 지체 없이 삭제합니다.
          이용자는 고객센터를 통해 열람/정정/삭제를 요청할 수 있습니다.
        </Text>
      </div>
      <LegalFooter />
    </div>
  );
}
