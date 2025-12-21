import React from "react";
import { Text } from "@toss-design-system/mobile";
import TopBar from "../components/TopBar";
import LegalFooter from "../components/LegalFooter";
import { APP_DISPLAY_NAME } from "../lib/env";

export default function HelpPage() {
  return (
    <div style={{ paddingBottom: 24 }}>
      <TopBar title="이용안내" />
      <div style={{ padding: 16 }}>
        <Text typography="t6" fontWeight="bold">서비스 개요</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          {APP_DISPLAY_NAME}은(는) 사용자가 선택한 챌린지를 수행하고, 정해진 인증 조건을 충족하면 리워드를 제공하는 “습관 관리” 서비스입니다.
          사행성/투자 상품이 아니며, 재산 증식을 보장하지 않습니다.
        </Text>

        <div style={{ height: 16 }} />
        <Text typography="t6" fontWeight="bold">이용 흐름</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          (1) 챌린지 선택 → (2) 참가비 결제 → (3) 인증 제출 → (4) 기간 종료 후 결과 확인/리워드 제공
        </Text>

        <div style={{ height: 16 }} />
        <Text typography="t6" fontWeight="bold">이용 제한</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          앱인토스(WebView) 테스트/이용은 토스 정책에 따라 워크스페이스 멤버 여부 및 연령 조건 등으로 제한될 수 있습니다.
        </Text>

        <div style={{ height: 16 }} />
        <Text typography="t6" fontWeight="bold">부정 사용 방지</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          동일 이미지 재사용 차단, 촬영시간(EXIF) 검증, 자동 분류(비전) 1차 필터, 신뢰 데이터(만보기/SDK) 연동을 순차 적용합니다.
        </Text>
      </div>
      <LegalFooter />
    </div>
  );
}
