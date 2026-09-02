const assert = require("node:assert/strict");

const baseURL = process.env.CLIPLI_BASE_URL || "http://127.0.0.1:4173";
const adminKey = process.env.CLIPLI_TEST_ADMIN_KEY || "local-comprehensive-admin";
const executorKey = process.env.CLIPLI_TEST_EXECUTOR_KEY || "local-comprehensive-executor";
const wallet = "0x1111111111111111111111111111111111111111";

async function request(path, { method = "GET", body, headers = {}, status = 200 } = {}) {
  const response = await fetch(baseURL + path, {
    method,
    headers: { ...(body === undefined ? {} : { "Content-Type": "application/json" }), ...headers },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await response.text();
  let payload = null;
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      assert.fail(`${method} ${path} returned non-JSON: ${text.slice(0, 160)}`);
    }
  }
  assert.equal(response.status, status, `${method} ${path}: ${text}`);
  return { response, payload };
}

async function main() {
  const results = [];
  const check = async (name, fn) => {
    await fn();
    results.push(name);
  };

  await check("health and security headers", async () => {
    const { response, payload } = await request("/api/health");
    assert.equal(payload.data.status, "ok");
    assert.match(response.headers.get("content-security-policy") || "", /frame-ancestors 'none'/);
    assert.equal(response.headers.get("x-content-type-options"), "nosniff");
    assert.equal(response.headers.get("cache-control"), "no-store");
  });

  await check("public portfolio and contract-shaped reads", async () => {
    for (const path of [
      "/api/v1/session",
      "/api/v1/hapw/summary",
      "/api/v1/hapw/assets",
      "/api/v1/hapw/assets/asset-2048",
      "/api/v1/hapw/assets/asset-2048/authorization",
      "/api/v1/hapw/assets/asset-2048/history",
      "/api/v1/clip/distribution-rules",
      "/api/v1/airdrop-rules",
    ]) {
      await request(path);
    }
  });

  await check("BNB wallet validation and connection", async () => {
    await request("/api/v1/wallet/connect", {
      method: "POST",
      body: { provider: "MetaMask", address: wallet, chainId: "0x1" },
      status: 400,
    });
    const { payload } = await request("/api/v1/wallet/connect", {
      method: "POST",
      body: { provider: "MetaMask", address: wallet, chainId: "0x61" },
    });
    assert.equal(payload.data.wallet, wallet);
    assert.equal(payload.data.walletChainId, "0x61");
  });

  await check("redemption, automatic entitlement and idempotency", async () => {
    const input = { requestId: "live-e2e-redeem-2048", assetId: "asset-2048", accepted: true };
    const { payload } = await request("/api/v1/hapw/redemptions", { method: "POST", body: input, status: 201 });
    assert.equal(payload.data.redemption.assetId, "asset-2048");
    assert.equal(payload.data.redemption.clipGranted, 60);
    await request("/api/v1/hapw/redemptions", { method: "POST", body: input });
    const conflict = await request("/api/v1/hapw/redemptions", {
      method: "POST",
      body: { ...input, assetId: "asset-771" },
      status: 409,
    });
    assert.equal(conflict.payload.error.code, "request_id_reused");
    const airdrops = await request(`/api/v1/airdrops?address=${wallet}&status=queued`);
    assert.ok(airdrops.payload.data.items.some((item) => item.assetId === "asset-2048" && item.amount === 60));
  });

  await check("generation, license conversion, platform exercise and reserve exchange", async () => {
    const operations = [
      ["/api/generations", { requestId: "live-e2e-generation-2048", assetId: "asset-2048", duration: 15, quality: "standard", accepted: true }],
      ["/api/conversions", { requestId: "live-e2e-conversion-771", assetId: "asset-771", region: "EU", days: 30, accepted: true }],
      ["/api/v1/hapw/exercises", { requestId: "live-e2e-exercise-332", assetId: "asset-332", platformCode: "foundation", accepted: true }],
      ["/api/v1/hapw/exchanges", { requestId: "live-e2e-exchange-528", assetId: "asset-528", accepted: true }],
    ];
    for (const [path, body] of operations) {
      await request(path, { method: "POST", body, status: 201 });
      await request(path, { method: "POST", body });
      const changed = { ...body, accepted: false };
      const conflict = await request(path, { method: "POST", body: changed, status: 409 });
      assert.equal(conflict.payload.error.code, "request_id_reused");
    }
  });

  await check("admin and executor authorization, simulation and terminal states", async () => {
    const adminHeaders = { "X-Clipli-Admin-Key": adminKey };
    const executorHeaders = { "X-Clipli-Executor-Key": executorKey };
    await request("/api/v1/admin/bnb-networks", { status: 401 });
    const networks = await request("/api/v1/admin/bnb-networks", { headers: adminHeaders });
    assert.equal(networks.payload.data.simulationOnly, true);

    const createBody = { requestId: "live-e2e-airdrop-sim", ruleCode: "admin_approved", walletAddress: wallet, chainId: "0x61", amount: 12 };
    const created = await request("/api/v1/admin/airdrops", { method: "POST", body: createBody, headers: adminHeaders, status: 201 });
    await request("/api/v1/admin/airdrops", { method: "POST", body: createBody, headers: adminHeaders });
    const conflict = await request("/api/v1/admin/airdrops", { method: "POST", body: { ...createBody, amount: 13 }, headers: adminHeaders, status: 409 });
    assert.equal(conflict.payload.error.code, "request_id_reused");
    const simulated = await request(`/api/v1/admin/airdrops/${created.payload.data.id}/simulate`, { method: "POST", body: { chainId: "0x61" }, headers: adminHeaders });
    assert.equal(simulated.payload.data.executionMode, "simulated-bnb");

    const executorItem = await request("/api/v1/admin/airdrops", {
      method: "POST",
      body: { ...createBody, requestId: "live-e2e-airdrop-executor", amount: 7 },
      headers: adminHeaders,
      status: 201,
    });
    await request(`/api/v1/internal/airdrops/${executorItem.payload.data.id}/result`, {
      method: "POST",
      body: { status: "confirmed", txHash: "0xabc", executorRef: "e2e" },
      headers: executorHeaders,
      status: 400,
    });
    const confirmed = await request(`/api/v1/internal/airdrops/${executorItem.payload.data.id}/result`, {
      method: "POST",
      body: { status: "confirmed", txHash: `0x${"b".repeat(64)}`, executorRef: "e2e-local-receipt" },
      headers: executorHeaders,
    });
    assert.equal(confirmed.payload.data.status, "confirmed");

    const failedItem = await request("/api/v1/admin/airdrops", {
      method: "POST",
      body: { ...createBody, requestId: "live-e2e-airdrop-failed", amount: 5 },
      headers: adminHeaders,
      status: 201,
    });
    await request(`/api/v1/internal/airdrops/${failedItem.payload.data.id}/result`, {
      method: "POST",
      body: { status: "failed", failureReason: "e2e simulated executor rejection" },
      headers: executorHeaders,
    });
    const terminal = await request(`/api/v1/internal/airdrops/${failedItem.payload.data.id}/result`, {
      method: "POST",
      body: { status: "confirmed", txHash: `0x${"c".repeat(64)}` },
      headers: executorHeaders,
      status: 409,
    });
    assert.equal(terminal.payload.error.code, "airdrop_failed_terminal");
  });

  await check("external binding boundary and treasury conservation", async () => {
    const unbound = await request("/api/v1/integrations/platform/users/e2e-unbound-user/assets", { status: 409 });
    assert.equal(unbound.payload.error.code, "user_not_bound");
    const treasury = await request("/api/v1/clip/treasury");
    assert.equal(treasury.payload.data.ledgerConservation.conserved, true);
  });

  console.log(JSON.stringify({ status: "passed", baseURL, checks: results.length, results }, null, 2));
}

main().catch((error) => {
  console.error(error.stack || error);
  process.exitCode = 1;
});
