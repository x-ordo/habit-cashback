import React from "react";
import { Text } from "@toss-design-system/mobile";
import { formatKRW } from "../lib/challenges";

export default function ReceiptCard({
  title,
  days,
  deposit,
}: {
  title: string;
  days: number;
  deposit: number;
}) {
  return (
    <div
      style={{
        background: "white",
        borderRadius: 16,
        padding: 16,
        boxShadow: "0 1px 6px rgba(0,0,0,0.06)",
      }}
    >
      <Text typography="t5" fontWeight="bold">
        정산 계약서
      </Text>

      <div style={{ height: 12 }} />

      <Row k="품목" v={title} />
      <Row k="기간" v={`${days}일 (작심삼일)`} />
      <Row k="참가비" v={formatKRW(deposit)} />

      <div style={{ height: 12, borderTop: "1px dashed #e5e8eb" }} />

      <Row k="성공 시" v={`${formatKRW(deposit)} 환급(리워드)`} />
      <Row k="실패 시" v="리워드 미지급 (정책 기준)" />

      <div style={{ height: 8 }} />
      <Text typography="t7" color="grey700">
        * 정산/리워드 지급 기준은 이용안내·약관을 따릅니다.
      </Text>
    </div>
  );
}

function Row({ k, v }: { k: string; v: string }) {
  return (
    <div style={{ display: "flex", justifyContent: "space-between", margin: "8px 0" }}>
      <Text typography="t6" color="grey700">
        {k}
      </Text>
      <Text typography="t6" fontWeight="medium">
        {v}
      </Text>
    </div>
  );
}
