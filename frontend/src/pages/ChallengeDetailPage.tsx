import React, { useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import TopBar from "../components/TopBar";
import ReceiptCard from "../components/ReceiptCard";
import BottomCTA from "../components/BottomCTA";
import LegalFooter from "../components/LegalFooter";
import { apiPost } from "../lib/api";
import { Challenge, OFFICIAL_CHALLENGES } from "../lib/challenges";

export default function ChallengeDetailPage() {
  const { id } = useParams();
  const nav = useNavigate();
  const [loading, setLoading] = useState(false);

  const challenge: Challenge | undefined = useMemo(
    () => OFFICIAL_CHALLENGES.find((c) => c.id === id),
    [id]
  );

  if (!challenge) {
    return (
      <div>
        <TopBar title="챌린지" />
        <div style={{ padding: 16 }}>존재하지 않는 챌린지입니다.</div>
      </div>
    );
  }

  const start = async () => {
    setLoading(true);
    try {
      const idem = crypto.randomUUID();
      const res = await apiPost<{ paymentId: string }>("/v1/payments/create", {
        challengeId: challenge.id,
        amount: challenge.deposit,
      }, idem);

      // In real: TossPay / 결제 SDK로 이어짐
      await apiPost("/v1/payments/execute", { paymentId: res.paymentId });

      nav(`/proof/${challenge.id}`, { replace: false });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ paddingBottom: 96 }}>
      <TopBar title="정산 계약" />
      <div style={{ padding: 16 }}>
        <ReceiptCard title={challenge.title} days={challenge.days} deposit={challenge.deposit} />
      </div>
      <BottomCTA label={loading ? "처리 중..." : `${challenge.deposit.toLocaleString()}원 보증금 맡기기`} onClick={start} disabled={loading} />
          <LegalFooter />
    </div>
  );
}
