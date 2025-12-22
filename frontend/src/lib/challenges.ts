export type ProofType = "photo" | "steps";

export type Challenge = {
  id: string;
  title: string;
  days: number;
  deposit: number;
  proofType: ProofType;
};

export const OFFICIAL_CHALLENGES: Challenge[] = [
  { id: "walk-7000", title: "매일 7,000보 걷기", days: 3, deposit: 10000, proofType: "steps" },
  { id: "bed-0700", title: "아침 7시 이불 개기", days: 3, deposit: 10000, proofType: "photo" },
  { id: "lunch-proof", title: "점심 도시락/샐러드 인증", days: 3, deposit: 10000, proofType: "photo" }
];

export function formatKRW(n: number) {
  return new Intl.NumberFormat("ko-KR").format(n) + "원";
}
