import React from "react";
import { Text } from "@toss-design-system/mobile";
import TopBar from "../components/TopBar";
import LegalFooter from "../components/LegalFooter";
import { APP_DISPLAY_NAME } from "../lib/env";

export default function TermsPage() {
  return (
    <div style={{ paddingBottom: 24 }}>
      <TopBar title="이용약관" />
      <div style={{ padding: 16 }}>
        <Text typography="t6" fontWeight="bold">제1조 (서비스의 성격)</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          {APP_DISPLAY_NAME}은(는) 사용자의 습관 형성을 돕기 위해 챌린지 참여, 인증, 결과 안내 및 리워드 제공 기능을 제공합니다.
          본 서비스는 사행성/투자 상품이 아니며, 수익을 보장하지 않습니다.
        </Text>

        <div style={{ height: 16 }} />
        <Text typography="t6" fontWeight="bold">제2조 (참가비 및 리워드)</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          참가비는 챌린지 참여를 위한 비용이며, 리워드는 성공 조건 충족 시 제공되는 혜택입니다.
          리워드의 형태(포인트/쿠폰 등), 기준 및 시점은 서비스 화면 및 공지에 따릅니다.
        </Text>

        <div style={{ height: 16 }} />
        <Text typography="t6" fontWeight="bold">제3조 (인증 및 제한)</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          인증 조건을 충족하지 못한 경우 리워드가 제공되지 않을 수 있습니다.
          부정행위(타인 사진/재사용 이미지/조작 등) 탐지 시 참여 제한 또는 계정 이용이 제한될 수 있습니다.
        </Text>

        <div style={{ height: 16 }} />
        <Text typography="t6" fontWeight="bold">제4조 (문의 및 분쟁)</Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          이용 관련 문의는 고객센터를 통해 접수할 수 있으며, 분쟁 발생 시 합리적인 범위에서 조정 절차를 진행합니다.
        </Text>
      </div>
      <LegalFooter />
    </div>
  );
}
