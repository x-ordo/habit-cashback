import React from "react";
import { Text } from "@toss-design-system/mobile";
import TopBar from "../components/TopBar";
import LegalFooter from "../components/LegalFooter";
import { SUPPORT_EMAIL, SUPPORT_HOURS } from "../lib/env";

export default function SupportPage() {
  return (
    <div style={{ paddingBottom: 24 }}>
      <TopBar title="고객센터" />
      <div style={{ padding: 16 }}>
        <Text typography="t6" fontWeight="bold">연락 채널</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">이메일: {SUPPORT_EMAIL}</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">운영시간: {SUPPORT_HOURS}</Text>

        <div style={{ height: 16 }} />
        <Text typography="t6" fontWeight="bold">문의 유형</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          1) 인증 실패/지연  2) 리워드 지급 문의  3) 부정사용 신고  4) 계정/접속 문제
        </Text>
      </div>
      <LegalFooter />
    </div>
  );
}
