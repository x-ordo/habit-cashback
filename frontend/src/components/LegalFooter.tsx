import React from "react";
import { Link } from "react-router-dom";
import { Text } from "@toss-design-system/mobile";

export default function LegalFooter() {
  return (
    <div style={{ padding: 16, paddingBottom: 24 }}>
      <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
        <Link to="/terms" style={{ textDecoration: "none" }}>
          <Text typography="t7" color="grey600">이용약관</Text>
        </Link>
        <Link to="/privacy" style={{ textDecoration: "none" }}>
          <Text typography="t7" color="grey600">개인정보처리방침</Text>
        </Link>
        <Link to="/support" style={{ textDecoration: "none" }}>
          <Text typography="t7" color="grey600">고객센터</Text>
        </Link>
        <Link to="/help" style={{ textDecoration: "none" }}>
          <Text typography="t7" color="grey600">이용안내</Text>
        </Link>
      </div>

      <div style={{ height: 10 }} />
      <Text typography="t8" color="grey500">
        * 앱인토스(WebView) 테스트/이용은 토스 정책에 따라 제한될 수 있습니다.
      </Text>
    </div>
  );
}
