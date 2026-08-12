const { works } = require('../../data');

async function listWorks() {
  return works;
}

module.exports = { listWorks };
