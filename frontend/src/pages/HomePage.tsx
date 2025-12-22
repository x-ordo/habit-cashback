import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { List, ListRow, Text } from "@toss-design-system/mobile";
import TopBar from "../components/TopBar";
import { APP_DISPLAY_NAME } from "../lib/env";
import LegalFooter from "../components/LegalFooter";
import { apiGet } from "../lib/api";
import { Challenge, formatKRW, OFFICIAL_CHALLENGES } from "../lib/challenges";

type Resp = { items: Challenge[] };

export default function HomePage() {
  const nav = useNavigate();
  const [items, setItems] = useState<Challenge[]>(OFFICIAL_CHALLENGES);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      try {
        const res = await apiGet<Resp>("/v1/challenges");
        if (res?.items?.length) setItems(res.items);
      } catch (e: any) {
        setErr(e?.message || null);
      }
    })();
  }, []);

  return (
    <div style={{ paddingBottom: 96 }}>
      <TopBar title={APP_DISPLAY_NAME} right={<a href="/help">이용안내</a>} />

      <div style={{ padding: "0 16px 16px" }}>
        <Text typography="t6" color="grey700">
          공식 챌린지 3개만 제공합니다. (심사/초기 운영 최적화)
        </Text>
        {err ? (
          <div style={{ marginTop: 8 }}>
            <Text typography="t7" color="grey600">
              백엔드 연결 실패: {err} (로컬 데모는 same-origin 기준)
            </Text>
          </div>
        ) : null}
      </div>

      <div style={{ background: "white" }}>
        <List>
          {items.map((c) => (
            <ListRow
              key={c.id}
              contents={
                <ListRow.Texts
                  type="2RowTypeA"
                  top={c.title}
                  bottom={`${c.days}일 · 참가비 ${formatKRW(c.deposit)}`}
                />
              }
              withArrow={true}
              onClick={() => nav(`/challenge/${c.id}`)}
            />
          ))}
        </List>
      </div>

      <div style={{ padding: 16 }}>
        <a href="/history" style={{ textDecoration: "none" }}>내 정산내역 보기 →</a>
      </div>
          <LegalFooter />
    </div>
  );
}
