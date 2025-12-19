import { tossLogin, tossCheckoutPayment } from './toss';

const API_BASE = import.meta.env.VITE_API_BASE ?? 'http://localhost:8080';

async function api(path: string, token: string | null, body?: any) {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw new Error(await res.text());
  return await res.json();
}

export async function runDemo() {
  // 1) Login via Toss SDK
  const { authorizationCode, referrer } = await tossLogin();

  // 2) Exchange authCode on backend -> get app token
  const auth = await api('/v1/auth/exchange', null, { authorizationCode, referrer });
  const token: string = auth.token;

  // 3) Create payment on backend -> get payToken
  const pay = await api('/v1/payments/create', token, {
    productDesc: '습관환급 참가비',
    amount: 10000,
    isTestPayment: true,
  });

  // 4) Show Toss payment UI
  const checkout = await tossCheckoutPayment(pay.payToken);
  if (checkout.resultType !== 'SUCCESS') {
    console.log('checkout failed', checkout);
    return;
  }

  // 5) Execute payment on backend
  const exec = await api('/v1/payments/execute', token, {
    orderNo: pay.orderNo,
    payToken: pay.payToken,
    isTestPayment: true,
  });

  console.log('payment executed', exec);
}
