import { appLogin, checkoutPayment } from '@apps-in-toss/web-framework';

export async function tossLogin() {
  // returns { authorizationCode, referrer }
  return await appLogin();
}

export async function tossCheckoutPayment(payToken: string) {
  // returns { resultType, success?, error? }
  return await checkoutPayment({ payToken });
}
