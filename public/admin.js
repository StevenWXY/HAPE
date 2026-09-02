(() => {
  const copy = {
    zh: {
      clearSession: "清除会话", eyebrow: "CLIPLI OPERATIONS", title: "空投运营后台", lead: "管理指定钱包和规则账户的 CLIP 空投任务。正式合约上线前可在 BNB 网络上模拟执行。", notConnected: "未连接运营接口", connected: "运营接口已连接", authEyebrow: "ACCESS", authTitle: "连接运营接口", authLead: "密钥只保存在当前浏览器会话中，不会写入 URL。", adminKey: "管理员密钥", executorKey: "执行器回调密钥（可选）", connect: "连接", createEyebrow: "AIRDROP CONTROL", createTitle: "创建空投任务", createLead: "可以按已满足的规则创建，也可以向指定钱包发放 CLIP。", protected: "受保护接口", rule: "空投规则", wallet: "目标钱包地址", chain: "BNB 网络", asset: "关联 HAPW（规则空投需要）", amount: "空投数量（CLIP）", requestId: "请求编号", check: "检查资格", queue: "创建任务", queueEyebrow: "EXECUTION QUEUE", queueTitle: "空投任务队列", queueLead: "查看已创建任务；尚未部署合约时可用模拟执行验证账本流程。", refresh: "刷新", statusFilter: "状态", allStatuses: "全部", walletFilter: "钱包筛选", filter: "筛选", tableWallet: "钱包", tableRule: "规则", tableAmount: "数量", tableStatus: "状态", tableCreated: "创建时间", tableTx: "回执", tableAction: "操作", simulate: "模拟执行", simulated: "模拟回执", emptyQueue: "暂无空投任务", rulesEyebrow: "RULES", rulesTitle: "当前规则", executorEyebrow: "EXECUTOR", executorTitle: "写入执行结果", executorLead: "正式链上广播由外部执行器完成；模拟执行不会访问私钥或发送交易。", airdropId: "空投 ID", resultStatus: "结果状态", txHash: "交易哈希", failureReason: "失败原因", writeResult: "写入结果", loading: "加载中…", noRules: "暂无可用规则", noAssets: "暂无 HAPW 资产", eligible: "满足资格", notEligible: "暂不满足", queued: "任务已创建", resultWritten: "执行结果已写入", invalidKey: "密钥无效或后台未配置，请检查服务端环境变量。", required: "请完整填写必填项。", publicExecutor: "执行器未配置或密钥无效。", treasury: "金库可分发", outstanding: "账本待结算", queuedCount: "排队任务", confirmedCount: "已确认", minted: "一次性铸造总量", source: "来源", manualRule: "管理员指定", redemptionRule: "HAPW 核销", noTx: "等待执行器", checkFailed: "资格检查失败", createFailed: "创建失败", refreshFailed: "刷新失败", updated: "已更新", statusFailed: "结果写入失败", simulationFailed: "模拟执行失败", simulationNote: "当前仅更新平台账本，不产生链上交易。"
    },
    en: {
      clearSession: "Clear session", eyebrow: "CLIPLI OPERATIONS", title: "Airdrop operations", lead: "Manage CLIP airdrops for specified wallets and rule-qualified accounts. Simulate BNB execution until the contract is deployed.", notConnected: "Operator API not connected", connected: "Operator API connected", authEyebrow: "ACCESS", authTitle: "Connect operator API", authLead: "Keys stay in this browser session and never enter the URL.", adminKey: "Admin key", executorKey: "Executor callback key (optional)", connect: "Connect", createEyebrow: "AIRDROP CONTROL", createTitle: "Create an airdrop", createLead: "Create a rule-qualified task or send CLIP to a specified wallet.", protected: "Protected API", rule: "Airdrop rule", wallet: "Target wallet", chain: "BNB network", asset: "Linked HAPW (required by rule)", amount: "Amount (CLIP)", requestId: "Request ID", check: "Check eligibility", queue: "Create task", queueEyebrow: "EXECUTION QUEUE", queueTitle: "Airdrop queue", queueLead: "Review tasks and executor status; simulation validates the ledger flow before contract deployment.", refresh: "Refresh", statusFilter: "Status", allStatuses: "All", walletFilter: "Wallet filter", filter: "Filter", tableWallet: "Wallet", tableRule: "Rule", tableAmount: "Amount", tableStatus: "Status", tableCreated: "Created", tableTx: "Receipt", tableAction: "Action", simulate: "Simulate", simulated: "Simulated receipt", emptyQueue: "No airdrop tasks", rulesEyebrow: "RULES", rulesTitle: "Active rules", executorEyebrow: "EXECUTOR", executorTitle: "Write execution result", executorLead: "Write back only after the external executor broadcasts; simulation never accesses a private key.", airdropId: "Airdrop ID", resultStatus: "Result status", txHash: "Transaction hash", failureReason: "Failure reason", writeResult: "Write result", loading: "Loading…", noRules: "No rules available", noAssets: "No HAPW assets", eligible: "Eligible", notEligible: "Not eligible", queued: "Airdrop queued", resultWritten: "Execution result written", invalidKey: "Invalid key or backend is not configured. Check the server environment.", required: "Complete all required fields.", publicExecutor: "Executor is not configured or the key is invalid.", treasury: "Treasury available", outstanding: "Ledger outstanding", queuedCount: "Queued", confirmedCount: "Confirmed", minted: "One-time minted supply", source: "Source", manualRule: "Admin specified", redemptionRule: "HAPW redemption", noTx: "Awaiting executor", checkFailed: "Eligibility check failed", createFailed: "Creation failed", refreshFailed: "Refresh failed", updated: "Updated", statusFailed: "Result update failed", simulationFailed: "Simulation failed", simulationNote: "This updates the platform ledger only; no blockchain transaction is created."
    }
  };
  const state = { lang: sessionStorage.getItem("clipli-admin-lang") || "zh", key: "", executorKey: "", rules: [], assets: [], queue: [] };
  const $ = (id) => document.getElementById(id);
  const text = (key) => (copy[state.lang] && copy[state.lang][key]) || copy.zh[key] || key;
  const escapeHTML = (value) => String(value == null ? "" : value).replace(/[&<>"']/g, char => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;", "'":"&#39;"}[char]));
  const shortAddress = (value) => value && value.length > 14 ? value.slice(0, 8) + "…" + value.slice(-6) : value || "-";
  const setFeedback = (id, message, kind = "") => { const node = $(id); if (!node) return; node.textContent = message || ""; node.className = "inline-feedback" + (kind ? " " + kind : ""); };

  function applyCopy() {
    document.documentElement.lang = state.lang === "en" ? "en" : "zh-CN";
    document.querySelectorAll("[data-copy]").forEach(node => { node.textContent = text(node.dataset.copy); });
    $("language-toggle").textContent = state.lang === "zh" ? "EN" : "中文";
  }

  async function request(path, options = {}, auth = "admin") {
    const headers = Object.assign({ "Accept": "application/json" }, options.headers || {});
    if (options.body) headers["Content-Type"] = "application/json";
    if (auth === "admin" && state.key) headers["X-Clipli-Admin-Key"] = state.key;
    if (auth === "executor" && state.executorKey) headers["X-Clipli-Executor-Key"] = state.executorKey;
    const response = await fetch(path, Object.assign({}, options, { headers }));
    const payload = await response.json().catch(() => ({}));
    if (!response.ok) {
      const error = payload.error || {};
      const err = new Error(error.message || "Request failed");
      err.code = error.code || "request_failed";
      err.status = response.status;
      throw err;
    }
    return payload.data;
  }

  function setConnection(connected, message, error = false) {
    const node = $("connection-state");
    node.className = "status-box" + (connected ? " connected" : error ? " error" : "");
    node.innerHTML = '<span class="status-dot"></span><span>' + escapeHTML(message) + '</span>';
  }

  function renderStats(treasury, queue) {
    const queued = queue.filter(item => item.status === "queued" || item.status === "submitted").length;
    const confirmed = queue.filter(item => item.status === "confirmed").length;
    $("stats").innerHTML = [
      [text("treasury"), treasury.treasuryBalance, "CLIP"],
      [text("outstanding"), treasury.ledgerOutstanding, "CLIP"],
      [text("queuedCount"), queued, text("source")],
      [text("confirmedCount"), confirmed, text("source")],
      [text("minted"), treasury.mintedSupply, "CLIP"]
    ].map(item => '<article class="stat-card"><small>' + escapeHTML(item[0]) + '</small><strong>' + Number(item[1] || 0).toLocaleString() + '</strong><span>' + escapeHTML(item[2]) + '</span></article>').join("");
    $("stats").style.gridTemplateColumns = "repeat(" + Math.min(5, Math.max(1, 5)) + ", minmax(0, 1fr))";
  }

  function renderRules() {
    $("rules-list").innerHTML = state.rules.length ? state.rules.map(rule => '<article class="rule-row"><span class="rule-code">' + escapeHTML(rule.code) + '</span><strong>' + escapeHTML(rule.name) + '</strong><small>' + escapeHTML(rule.description) + '</small><div class="rule-meta"><span>' + escapeHTML(rule.formula) + '</span><span>' + (rule.code === "admin_approved" ? text("manualRule") : text("redemptionRule")) + '</span></div></article>').join("") : '<p class="inline-feedback">' + text("noRules") + '</p>';
    $("rule-code").innerHTML = state.rules.map(rule => '<option value="' + escapeHTML(rule.code) + '">' + escapeHTML(rule.code) + ' · ' + escapeHTML(rule.name) + '</option>').join("");
    updateRuleFields();
  }

  function renderAssets() {
    $("asset-id").innerHTML = '<option value="">' + escapeHTML(text("noAssets")) + '</option>' + state.assets.map(asset => '<option value="' + escapeHTML(asset.id) + '">' + escapeHTML(asset.tokenId) + ' · ' + escapeHTML(asset.name) + ' · ' + escapeHTML(asset.redemptionStatus) + '</option>').join("");
  }

  function renderQueue() {
    const body = $("queue-body");
    if (!state.queue.length) { body.innerHTML = '<tr><td colspan="7" class="table-empty">' + text("emptyQueue") + '</td></tr>'; return; }
    body.innerHTML = state.queue.map(item => '<tr><td><strong title="' + escapeHTML(item.walletAddress) + '">' + escapeHTML(shortAddress(item.walletAddress)) + '</strong><small>' + escapeHTML(item.chainId) + '</small></td><td>' + escapeHTML(item.ruleCode) + '<small>' + escapeHTML(item.assetId || "-") + '</small></td><td><strong>' + Number(item.amount || 0).toLocaleString() + ' ' + escapeHTML(item.token) + '</strong></td><td><span class="status-pill ' + escapeHTML(item.status) + '">' + escapeHTML(item.status) + '</span>' + (item.simulated ? '<small>' + escapeHTML(text("simulated")) + '</small>' : '') + '</td><td>' + escapeHTML(item.createdAt) + '</td><td>' + (item.txHash ? '<code>' + escapeHTML(item.txHash) + '</code>' : '<span class="muted-cell">' + text("noTx") + '</span>') + '</td><td>' + ((item.status === "queued" || item.status === "submitted") ? '<button type="button" class="secondary-button queue-simulate" data-airdrop-id="' + escapeHTML(item.id) + '" data-chain-id="' + escapeHTML(item.chainId || "0x61") + '">' + escapeHTML(text("simulate")) + '</button>' : '<span class="muted-cell">-</span>') + '</td></tr>').join("");
    document.querySelectorAll(".queue-simulate").forEach(button => button.addEventListener("click", () => simulateAirdrop(button)));
  }

  function updateRuleFields() {
    const manual = $("rule-code").value === "admin_approved";
    $("amount-field").hidden = !manual;
    $("amount").disabled = !manual;
    $("asset-field").hidden = manual;
    $("asset-id").required = !manual;
  }

  async function loadAssets() {
    const data = await request("/api/v1/hapw/assets?page=1&pageSize=100", {}, "public");
    state.assets = data.items || [];
    renderAssets();
  }

  async function loadQueue() {
    const params = new URLSearchParams();
    const status = $("status-filter").value;
    const wallet = $("wallet-filter").value.trim();
    if (status) params.set("status", status);
    if (wallet) params.set("walletAddress", wallet);
    const data = await request("/api/v1/admin/airdrops" + (params.toString() ? "?" + params.toString() : ""));
    state.queue = data.items || [];
    renderQueue();
    return state.queue;
  }

  async function loadDashboard() {
    const [rules, treasury, networks] = await Promise.all([request("/api/v1/admin/airdrop-rules"), request("/api/v1/clip/treasury", {}, "public"), request("/api/v1/admin/bnb-networks")]);
    const options = (networks.items || []).filter(item => item.simulationOnly).map(item => '<option value="' + escapeHTML(item.chainId) + '">' + escapeHTML(item.name) + ' · ' + escapeHTML(item.nativeCurrency) + ' · ' + escapeHTML(state.lang === "zh" ? "模拟" : "simulation") + '</option>').join("");
    if (options) $("chain-id").innerHTML = options;
    state.rules = rules.items || [];
    renderRules();
    await loadAssets();
    const queue = await loadQueue();
    renderStats(treasury.treasury, queue);
  }

  function showDashboard() { $("auth-panel").hidden = true; $("dashboard").hidden = false; setConnection(true, text("connected")); }

  function makeRequestID() { $("request-id").value = "admin-" + Date.now().toString(36) + "-" + Math.random().toString(36).slice(2, 7); }

  async function connect() {
    const key = $("admin-key").value.trim();
    if (!key) { setFeedback("auth-feedback", text("required"), "error"); return; }
    state.key = key;
    state.executorKey = $("executor-key").value.trim();
    setFeedback("auth-feedback", text("loading"));
    try { await loadDashboard(); showDashboard(); setFeedback("auth-feedback", ""); } catch (error) { setConnection(false, text("notConnected"), true); setFeedback("auth-feedback", text("invalidKey"), "error"); }
  }

  async function checkEligibility() {
    const address = $("wallet-address").value.trim();
    const rule = $("rule-code").value;
    const asset = $("asset-id").value;
    if (!address || !rule || (rule === "hapw_redemption" && !asset)) { setFeedback("create-feedback", text("required"), "error"); return; }
    const params = new URLSearchParams({ address, ruleCode: rule });
    if (asset) params.set("assetId", asset);
    try {
      const data = await request("/api/v1/wallet/airdrop-eligibility?" + params.toString(), {}, "public");
      const label = data.eligible ? text("eligible") : text("notEligible");
      $("eligibility").hidden = false;
      $("eligibility").innerHTML = '<strong>' + escapeHTML(label) + '</strong> · ' + escapeHTML(data.reason) + (data.amount ? ' · ' + Number(data.amount).toLocaleString() + ' CLIP' : '') + (data.existingAirdropId ? ' · ' + escapeHTML(data.existingAirdropId) : '');
      if (data.amount && rule === "hapw_redemption") $("amount").value = data.amount;
    } catch (error) { setFeedback("create-feedback", text("checkFailed") + " · " + error.message, "error"); }
  }

  async function createAirdrop(event) {
    event.preventDefault();
    const rule = $("rule-code").value;
    const body = { requestId: $("request-id").value.trim(), ruleCode: rule, walletAddress: $("wallet-address").value.trim(), chainId: $("chain-id").value.trim(), assetId: $("asset-id").value };
    const amount = Number($("amount").value);
    if (rule === "admin_approved") body.amount = amount;
    if (!body.requestId || !body.walletAddress || !body.chainId || (rule === "admin_approved" && (!amount || amount < 1)) || (rule === "hapw_redemption" && !body.assetId)) { setFeedback("create-feedback", text("required"), "error"); return; }
    try {
      await request("/api/v1/admin/airdrops", { method: "POST", body });
      setFeedback("create-feedback", text("queued"), "success");
      $("eligibility").hidden = true;
      makeRequestID();
      await loadDashboard();
    } catch (error) { setFeedback("create-feedback", text("createFailed") + " · " + error.message, "error"); }
  }

  async function writeResult(event) {
    event.preventDefault();
    const id = $("result-id").value.trim();
    if (!id || !state.executorKey) { setFeedback("result-feedback", text("publicExecutor"), "error"); return; }
    const body = { status: $("result-status").value, txHash: $("tx-hash").value.trim(), failureReason: $("failure-reason").value.trim() };
    try { await request("/api/v1/internal/airdrops/" + encodeURIComponent(id) + "/result", { method: "POST", body }, "executor"); setFeedback("result-feedback", text("resultWritten"), "success"); await loadQueue(); } catch (error) { setFeedback("result-feedback", text("statusFailed") + " · " + error.message, "error"); }
  }

  async function simulateAirdrop(button) {
    const id = button.dataset.airdropId;
    button.disabled = true;
    try {
      await request("/api/v1/admin/airdrops/" + encodeURIComponent(id) + "/simulate", { method: "POST", body: JSON.stringify({ chainId: button.dataset.chainId || "0x61" }) });
      setFeedback("create-feedback", text("simulationNote"), "success");
      await loadDashboard();
    } catch (error) {
      setFeedback("create-feedback", text("simulationFailed") + " · " + error.message, "error");
      button.disabled = false;
    }
  }

  function clearSession() { state.key = ""; state.executorKey = ""; $("dashboard").hidden = true; $("auth-panel").hidden = false; $("admin-key").value = ""; $("executor-key").value = ""; setConnection(false, text("notConnected")); setFeedback("auth-feedback", ""); }

  $("language-toggle").addEventListener("click", () => { state.lang = state.lang === "zh" ? "en" : "zh"; sessionStorage.setItem("clipli-admin-lang", state.lang); applyCopy(); if (!$("dashboard").hidden) { renderRules(); renderAssets(); renderQueue(); } });
  $("clear-session").addEventListener("click", clearSession);
  $("auth-form").addEventListener("submit", event => { event.preventDefault(); connect(); });
  $("airdrop-form").addEventListener("submit", createAirdrop);
  $("check-eligibility").addEventListener("click", checkEligibility);
  $("rule-code").addEventListener("change", updateRuleFields);
  $("refresh-queue").addEventListener("click", () => loadDashboard().catch(() => setFeedback("create-feedback", text("refreshFailed"), "error")));
  $("filter-queue").addEventListener("click", () => loadQueue().catch(error => setFeedback("create-feedback", text("refreshFailed") + " · " + error.message, "error")));
  $("result-form").addEventListener("submit", writeResult);
  makeRequestID();
  applyCopy();
})();
