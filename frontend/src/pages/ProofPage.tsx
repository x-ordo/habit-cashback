import React, { useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Button, Text } from "@toss-design-system/mobile";
import TopBar from "../components/TopBar";
import LegalFooter from "../components/LegalFooter";
import BottomCTA from "../components/BottomCTA";
import { apiPost } from "../lib/api";
import { OFFICIAL_CHALLENGES } from "../lib/challenges";

export default function ProofPage() {
  const { id } = useParams();
  const nav = useNavigate();
  const ch = useMemo(() => OFFICIAL_CHALLENGES.find((c) => c.id === id), [id]);

  const [file, setFile] = useState<File | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  if (!ch) {
    return (
      <div>
        <TopBar title="인증" />
        <div style={{ padding: 16 }}>챌린지를 찾을 수 없습니다.</div>
      </div>
    );
  }

  const submit = async () => {
    setSubmitting(true);
    setMsg(null);

    try {
      const idem = crypto.randomUUID();
      if (ch.proofType === "photo") {
        if (!file) throw new Error("사진을 선택하세요.");
        const base64 = await fileToBase64(file);
        await apiPost("/v1/proofs/submit", { challengeId: ch.id, imageBase64: base64 }, idem);
      } else {
        // steps proof: demo only
        await apiPost("/v1/proofs/submit", { challengeId: ch.id, imageHash: "steps-demo" }, idem);
      }

      setMsg("인증 완료. 정산 대기 중입니다.");
      setTimeout(() => nav("/history"), 500);
    } catch (e: any) {
      setMsg(e?.message || "인증 실패");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div style={{ paddingBottom: 120 }}>
      <TopBar title="오늘 인증" />

      <div style={{ padding: 16 }}>
        <Text typography="t6" fontWeight="bold">
          {ch.title}
        </Text>
        <div style={{ height: 8 }} />
        <Text typography="t7" color="grey700">
          {ch.proofType === "photo"
            ? "사진 인증(카메라 권장). EXIF/중복검증은 서버에서 처리(로드맵)."
            : "만보기/걸음수 인증(데모)."}
        </Text>

        <div style={{ height: 16 }} />

        {ch.proofType === "photo" ? (
          <div style={{ background: "white", borderRadius: 16, padding: 16 }}>
            <input
              type="file"
              accept="image/*"
              capture="environment"
              onChange={(e) => setFile(e.target.files?.[0] || null)}
            />
            {file ? (
              <div style={{ marginTop: 12 }}>
                <Text typography="t7" color="grey700">
                  선택됨: {file.name}
                </Text>
              </div>
            ) : null}
          </div>
        ) : (
          <div style={{ background: "white", borderRadius: 16, padding: 16 }}>
            <Text typography="t7" color="grey700">
              데모: 걸음수는 서버에서 임의로 승인됩니다.
            </Text>
          </div>
        )}

        <div style={{ height: 12 }} />
        {msg ? (
          <div style={{ background: "white", borderRadius: 12, padding: 12 }}>
            <Text typography="t7">{msg}</Text>
          </div>
        ) : null}

        <div style={{ height: 24 }} />
        <Button onClick={() => nav("/history")} style={{ width: "100%" }}>
          정산내역 보기
        </Button>
      </div>

      <BottomCTA label={submitting ? "제출 중..." : "인증 제출"} onClick={submit} disabled={submitting} />
          <LegalFooter />
    </div>
  );
}

async function fileToBase64(file: File): Promise<string> {
  const buf = await file.arrayBuffer();
  const bytes = new Uint8Array(buf);
  let binary = "";
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}
