// k6 2.4.0: one simulated player through a full 12-question show (stack/14-testing-quality).
// Scale: ~20 runners x 30k VUs = 600k players (k6-operator or EC2 fleet). Never run against prod.
// Env: BASE_URL (staging game ALB), SHOW_ID, TOKEN_POOL (pre-minted test tokens, DPoP disabled
// only in the dedicated load-test environment).
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

const BASE = __ENV.BASE_URL;
const SHOW = __ENV.SHOW_ID;
const QUESTIONS = 12;
const ackLatency = new Trend('answer_ack_ms', true);
const lostAnswers = new Counter('answers_not_acked');

const json = { headers: { 'Content-Type': 'application/json', 'Connect-Protocol-Version': '1' } };
const rpc = (method, body, extraHeaders = {}) =>
  http.post(`${BASE}/tirivia.game.v1.GameService/${method}`, JSON.stringify(body), {
    headers: { ...json.headers, Authorization: `Bearer ${__ENV.TOKEN}`, ...extraHeaders },
  });

export const options = {
  scenarios: {
    show: {
      executor: 'per-vu-iterations',
      vus: Number(__ENV.VUS || 30000),
      iterations: 1,
      maxDuration: '30m',
    },
  },
  thresholds: {
    answer_ack_ms: ['p(99)<300'],
    answers_not_acked: ['count==0'],
    http_req_failed: ['rate<0.001'],
  },
};

export default function () {
  // Lobby: join spread over the last few minutes before air.
  sleep(Math.random() * 180);
  const join = rpc('JoinShow', { show_id: SHOW });
  check(join, { joined: (r) => r.status === 200 });

  for (let q = 1; q <= QUESTIONS; q++) {
    // Wait for the question cue (driven by the rehearsal studio clock in the test env).
    sleep(25 + Math.random() * 2);
    // Answer burst shaped like a real timer: most answers in the first half of the window.
    sleep(Math.min(Math.random() ** 2 * 8, 8));
    const res = rpc(
      'SubmitAnswer',
      { show_id: SHOW, question_id: `${SHOW}-q${q}`, option_id: ['a', 'b', 'c'][Math.floor(Math.random() * 3)] },
      { 'Idempotency-Key': `"${__VU}-${q}"` },
    );
    ackLatency.add(res.timings.duration);
    if (res.status !== 200) lostAnswers.add(1);

    // Reactions: ~30% of players react, one batch per 5 s (+-1 s jitter), only when tapped (ADR-0006).
    if (Math.random() < 0.3) {
      rpc('SubmitReactions', { show_id: SHOW, counts: { 1: 1 + Math.floor(Math.random() * 5) } });
    }
  }
  // Reconnect storm sample: 5% of players fetch state after the final.
  if (Math.random() < 0.05) rpc('GetShowState', { show_id: SHOW });
}
