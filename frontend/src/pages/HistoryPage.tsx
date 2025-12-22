import React, { useEffect, useState } from "react";
import { List, ListRow, Text } from "@toss-design-system/mobile";
import TopBar from "../components/TopBar";
import LegalFooter from "../components/LegalFooter";
import { apiGet } from "../lib/api";
import { OFFICIAL_CHALLENGES } from "../lib/challenges";

type Settlement = {
  challengeId: string;
  status: "running" | "success" | "failed";
  refundable: boolean;
  message?: string;
};

export default function HistoryPage() {
  const [items, setItems] = useState<Settlement[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      try {
        const res: Settlement[] = [];
        for (const ch of OFFICIAL_CHALLENGES) {
          const s = await apiGet<Settlement>(`/v1/settlements/${ch.id}`);
          res.push(s);
        }
        setItems(res);
      } catch (e: any) {
        setErr(e?.message || null);
      }
    })();
  }, []);

  return (
    <div>
      <TopBar title="정산내역" />
      {err ? (
        <div style={{ padding: 16 }}>
          <Text typography="t7" color="grey700">
            불러오기 실패: {err}
          </Text>
        </div>
      ) : null}
      <div style={{ background: "white" }}>
        <List>
          {items.map((it) => (
            <ListRow
              key={it.challengeId}
              contents={
                <ListRow.Texts
                  type="2RowTypeA"
                  top={it.challengeId}
                  bottom={`${it.status} · refundable=${it.refundable ? "Y" : "N"}`}
                />
              }
            />
          ))}
        </List>
      </div>
          <LegalFooter />
    </div>
  );
}
