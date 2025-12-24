import React, { useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { pay } from "@apps-in-toss/web-framework";
import TopBar from "../components/TopBar";
import ReceiptCard from "../components/ReceiptCard";
import BottomCTA from "../components/BottomCTA";
import LegalFooter from "../components/LegalFooter";
import { apiPost } from "../lib/api";
import { Challenge, OFFICIAL_CHALLENGES } from "../lib/challenges";

interface PaymentCreateResponse {
  paymentId: number;
  orderNo: string;
  payToken: string;
  status: string;
  challengeId: string;
  amount: number;
  mode: "mock" | "live";
}

export default function ChallengeDetailPage() {
  const { id } = useParams();
  const nav = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
    setError(null);

    try {
      const idem = crypto.randomUUID();

      // 1. Create payment and get payToken
      const res = await apiPost<PaymentCreateResponse>("/v1/payments/create", {
        challengeId: challenge.id,
        amount: challenge.deposit,
      }, idem);

      // 2. Handle payment based on mode
      if (res.mode === "mock") {
        // Mock mode: Skip TossPay UI and execute directly
        await apiPost("/v1/payments/execute", { paymentId: res.paymentId });
        nav(`/proof/${challenge.id}`, { replace: false });
      } else {
        // Live mode: Call TossPay SDK
        try {
          const payResult = await pay.requestPayment({
            payToken: res.payToken,
          });

          if (payResult.success) {
            // Payment approved by user, execute on backend
            await apiPost("/v1/payments/execute", { paymentId: res.paymentId });
            nav(`/proof/${challenge.id}`, { replace: false });
          } else {
            // Payment cancelled or failed
            setError(payResult.errorMessage || "결제가 취소되었습니다.");
          }
        } catch (payError: unknown) {
          console.error("TossPay error:", payError);
          setError("결제 처리 중 오류가 발생했습니다.");
        }
      }
    } catch (e: unknown) {
      console.error("Payment error:", e);
      const errMsg = e instanceof Error ? e.message : "결제 생성 중 오류가 발생했습니다.";
      setError(errMsg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ paddingBottom: 96 }}>
      <TopBar title="정산 계약" />
      <div style={{ padding: 16 }}>
        <ReceiptCard title={challenge.title} days={challenge.days} deposit={challenge.deposit} />
        {error && (
          <div style={{
            marginTop: 16,
            padding: 12,
            backgroundColor: "#FEE2E2",
            color: "#DC2626",
            borderRadius: 8,
            fontSize: 14,
          }}>
            {error}
          </div>
        )}
      </div>
      <BottomCTA label={loading ? "처리 중..." : `${challenge.deposit.toLocaleString()}원 보증금 맡기기`} onClick={start} disabled={loading} />
      <LegalFooter />
    </div>
  );
}
