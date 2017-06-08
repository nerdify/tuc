const { send } = require('micro');
const { cache, createClient } = require('r-cache');
const { getBalance } = require('tuc-promise');

const client = createClient(
  process.env.REDIS_PORT, process.env.REDIS_HOST, { no_ready_check: true }
);

client.auth(process.env.REDIS_PASSWORD);

module.exports = async (req, res) => {
  const number = req.url.replace(/.*(\d{8}).*/, '$1');
  const cacheKey = `tuc:${number}`;

  try {
    const response = await cache(cacheKey, 120, () => getBalance(number));

    send(res, 200, {
      number,
      balance: parseFloat(response),
    });
  } catch (err) {
    const { code } = err;

    if (code === 100 || code === 104) {
      return send(res, 404);
    }

    send(res, 503);
  }
}
