import React, { useEffect, useState } from "react";
import { List, ListRow, Text } from "@toss-design-system/mobile";
import TopBar from "../components/TopBar";
import LegalFooter from "../components/LegalFooter";
import { apiGet } from "../lib/api";

type Settlement = {
  challengeId: string;
  status: "running" | "success" | "failed";
  refundable: boolean;
  message?: string;
};

type SettlementsResponse = {
  items: Settlement[];
};

const statusIcon = (status: string) => {
  switch (status) {
    case "running":
      return "⏳";
    case "success":
      return "✅";
    case "failed":
      return "❌";
    default:
      return "•";
  }
};

export default function HistoryPage() {
  const [items, setItems] = useState<Settlement[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      try {
        // Use batch API for better performance
        const res = await apiGet<SettlementsResponse>("/v1/settlements");
        setItems(res.items || []);
      } catch (e: any) {
        setErr(e?.message || "불러오기 실패");
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  return (
    <div>
      <TopBar title="정산내역" />

      {loading && (
        <div style={{ padding: 16, textAlign: "center" }}>
          <Text typography="t7" color="grey500">
            불러오는 중...
          </Text>
        </div>
      )}

      {err && !loading && (
        <div style={{ padding: 16 }}>
          <Text typography="t7" color="grey700">
            {err}
          </Text>
        </div>
      )}

      {!loading && !err && items.length === 0 && (
        <div style={{ padding: 32, textAlign: "center" }}>
          <Text typography="t6" color="grey500">
            아직 참여한 챌린지가 없습니다
          </Text>
        </div>
      )}

      {!loading && items.length > 0 && (
        <div style={{ background: "white" }}>
          <List>
            {items.map((it) => (
              <ListRow
                key={it.challengeId}
                contents={
                  <ListRow.Texts
                    type="2RowTypeA"
                    top={`${statusIcon(it.status)} ${it.challengeId}`}
                    bottom={it.message || `상태: ${it.status}`}
                  />
                }
                right={
                  it.refundable ? (
                    <Text typography="t7" color="blue500">
                      환급 예정
                    </Text>
                  ) : null
                }
              />
            ))}
          </List>
        </div>
      )}

      <LegalFooter />
    </div>
  );
}
