if (process.env.NODE_ENV !== 'production') {
	require('dotenv').config();
}

const micro = require('micro');
const {cache, createClient} = require('r-cache');
const {getBalance} = require('tuc-promise');

const {createError, send} = micro;
const client = createClient(
	// eslint-disable-next-line camelcase
	process.env.REDIS_PORT, process.env.REDIS_HOST, {no_ready_check: true}
);

if (process.env.REDIS_PASSWORD) {
	client.auth(process.env.REDIS_PASSWORD);
}

const server = micro(async (req, res) => {
	const number = req.url.replace(/.*(\d{8}).*/, '$1');
	const cacheKey = `tuc:${number}`;

	try {
		const response = await cache(cacheKey, 120, () => getBalance(number));

		send(res, 200, {
			number,
			balance: parseFloat(response)
		});
	} catch (err) {
		const {code = 503, message = ''} = err;

		if (code === 100) {
			return send(res, 404);
		}

		throw createError(code, message, err);
	}
});

server.listen(process.env.PORT || 4000);
