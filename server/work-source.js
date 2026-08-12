const seed = require('./adapters/works/seed');
const { normalizeWorks } = require('./adapters/works/normalize');

const sources = { seed };

function current() {
  const sourceName = process.env.WORK_SOURCE || 'seed';
  return sources[sourceName] || seed;
}

async function listWorks() {
  const works = await current().listWorks();
  return normalizeWorks(works);
}

module.exports = { listWorks };
