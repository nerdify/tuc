const Promise = require('bluebird');
const { send } = require('micro');
const Tuc = require('tuc');

const tuc = new Tuc;

module.exports = async (req, res) => {
  const number = req.url.replace(/.*(\d{8}).*/, '$1');

  const response = await new Promise((resolve) => {
    tuc.getBalance(number, (data) => resolve(data));
  });

  if (tuc.isError(response)) {
    const { code } = response.error;

    if (code === 100 || code === 104) {
      return send(res, 404);
    }

    return send(res, 503);
  }

  send(res, 200, {
    number,
    balance: parseFloat(response),
  });
}
