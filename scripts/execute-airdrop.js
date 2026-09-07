// Run one approved task. Persist the signed transaction before any broadcast
// so retries always rebroadcast identical bytes, including the original nonce.
const fs = require('node:fs');
const path = require('node:path');
const { ethers } = require('ethers');

async function main() {
  const id = process.argv[2];
  if (!id || !/^airdrop-[a-zA-Z0-9-]+$/.test(id)) throw new Error('Usage: node scripts/execute-airdrop.js <airdrop-id>');
  for (const key of ['CLIPLI_ADMIN_API_KEY', 'CLIPLI_AIRDROP_EXECUTOR_KEY', 'CLIPLI_BNB_RPC_URL', 'CLIPLI_CLIP_CONTRACT', 'CLIPLI_TREASURY_ADDRESS', 'CLIPLI_EXECUTOR_PRIVATE_KEY']) {
    if (!process.env[key]) throw new Error('Missing ' + key);
  }
  const base = process.env.CLIPLI_API_URL || 'http://127.0.0.1:4173';
  const directory = path.resolve(process.env.CLIPLI_EXECUTOR_JOURNAL || '.data/airdrop-executor');
  fs.mkdirSync(directory, {recursive: true, mode: 0o700});
  const lockPath = path.join(directory, 'executor.lock');
  const lock = fs.openSync(lockPath, 'wx', 0o600);
  try {
    async function api(route, body) {
      const response = await fetch(base + route, {method: body ? 'POST' : 'GET', headers: {'Content-Type': 'application/json', 'X-Clipli-Admin-Key': process.env.CLIPLI_ADMIN_API_KEY, 'X-Clipli-Executor-Key': process.env.CLIPLI_AIRDROP_EXECUTOR_KEY}, body: body ? JSON.stringify(body) : undefined, signal: AbortSignal.timeout(30000)});
      const result = await response.json();
      if (!response.ok) throw new Error(result.error?.code || 'Clipli API request failed');
      return result.data;
    }
    const tasks = await api('/api/v1/admin/airdrops');
    const task = tasks.items.find(item => item.id === id);
    if (!task || !['queued', 'submitted'].includes(task.status) || task.simulated || task.chainId !== '0x38') throw new Error('Task must be an active real BNB mainnet airdrop');
    const provider = new ethers.JsonRpcProvider(process.env.CLIPLI_BNB_RPC_URL);
    const signer = new ethers.Wallet(process.env.CLIPLI_EXECUTOR_PRIVATE_KEY, provider);
    if ((await provider.getNetwork()).chainId !== 56n || signer.address.toLowerCase() !== process.env.CLIPLI_TREASURY_ADDRESS.toLowerCase()) throw new Error('Wrong chain or treasury signer');
    const token = new ethers.Contract(process.env.CLIPLI_CLIP_CONTRACT, ['function decimals() view returns (uint8)', 'function balanceOf(address) view returns (uint256)', 'function transfer(address,uint256) returns (bool)'], signer);
    if (await token.decimals() !== 18n || !Number.isSafeInteger(task.amount) || task.amount <= 0) throw new Error('Invalid CLIP decimals or amount');
    const amount = ethers.parseUnits(String(task.amount), 18);
    const journalPath = path.join(directory, id + '.json');
    let journal;
    if (fs.existsSync(journalPath)) {
      journal = JSON.parse(fs.readFileSync(journalPath, 'utf8'));
      const signed = ethers.Transaction.from(journal.raw);
      const expected = token.interface.encodeFunctionData('transfer', [task.walletAddress, amount]);
      if (signed.hash !== journal.hash || signed.from?.toLowerCase() !== signer.address.toLowerCase() || signed.chainId !== 56n || signed.to?.toLowerCase() !== token.target.toLowerCase() || signed.data !== expected) throw new Error('Signed journal does not match this task');
    } else {
      if (task.status === 'submitted') throw new Error('Task already submitted; recover its original executor journal');
      if (await token.balanceOf(signer.address) < amount) throw new Error('Insufficient CLIP treasury balance');
      const transaction = await signer.populateTransaction(await token.transfer.populateTransaction(task.walletAddress, amount));
      const raw = await signer.signTransaction(transaction);
      journal = {id, hash: ethers.keccak256(raw), raw};
      const file = fs.openSync(journalPath, 'wx', 0o600);
      try { fs.writeFileSync(file, JSON.stringify(journal)); fs.fsyncSync(file); } finally { fs.closeSync(file); }
    }
    if (task.txHash && task.txHash !== journal.hash) throw new Error('Task has another transaction hash');
    const callback = '/api/v1/internal/airdrops/' + encodeURIComponent(id) + '/result';
    // The server rejects cancellation once the immutable hash is registered.
    await api(callback, {status: 'submitted', txHash: journal.hash, executorRef: 'bnb-erc20-executor'});
    const existing = await provider.getTransaction(journal.hash);
    if (!existing) await provider.broadcastTransaction(journal.raw);
    const receipt = await provider.waitForTransaction(journal.hash, 3, 120000);
    if (!receipt) { console.log('Transaction submitted: ' + journal.hash + '. Run the same command to resume confirmation.'); return; }
    await api(callback, {status: receipt.status === 1 ? 'confirmed' : 'failed', txHash: journal.hash, executorRef: 'bnb-erc20-executor', failureReason: receipt.status === 1 ? '' : 'On-chain transfer reverted'});
    console.log('Airdrop ' + id + ': ' + (receipt.status === 1 ? 'confirmed' : 'failed') + ' ' + journal.hash);
  } finally { fs.closeSync(lock); fs.unlinkSync(lockPath); }
}

main().catch(error => { console.error(error.shortMessage || error.message); process.exitCode = 1; });
