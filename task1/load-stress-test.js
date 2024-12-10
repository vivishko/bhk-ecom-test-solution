import http from "k6/http";
import { check } from "k6";
import { sleep } from "k6";

export const options = {
  stages: [
    { duration: "2m", target: 100 },
    { duration: "6m", target: 500 },
    { duration: "2m", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<200"],
    http_req_failed: ["rate<0.01"],
  },
};

export default function () {
  const bannerId = Math.floor(Math.random() * (100 - 10 + 1)) + 10;

  const url = `http://localhost:8080/counter/${bannerId}`;
  const payload = JSON.stringify({});
  const params = {
    headers: { "Content-Type": "application/json" },
  };

  const res = http.post(url, payload, params);

  check(res, {
    "status is 200": (r) => r.status === 200,
  });

  sleep(0.1);
}
