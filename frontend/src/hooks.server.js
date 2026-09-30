/**
 * Server-side hook to proxy /api and /health requests to the Go backend service in production.
 */
export async function handle({ event, resolve }) {
	const pathname = event.url.pathname;
	if (pathname.startsWith('/api') || pathname === '/health') {
		const backendHost = process.env.BACKEND_INTERNAL_URL || 'http://localhost:8080';
		const targetUrl = `${backendHost}${pathname}${event.url.search}`;

		// Forward request to Go backend with public host info
		const init = {
			method: event.request.method,
			headers: new Headers(event.request.headers)
		};

		const fwdProto = event.request.headers.get('x-forwarded-proto');
		const isLocal = event.url.hostname === 'localhost' || event.url.hostname === '127.0.0.1' || event.url.hostname.startsWith('192.168.') || event.url.hostname.startsWith('10.');
		const proto = fwdProto || (isLocal ? 'http' : (event.url.protocol.replace(':', '') || 'http'));

		init.headers.set('X-Forwarded-Host', event.url.host);
		init.headers.set('X-Forwarded-Proto', proto);

		if (event.request.method !== 'GET' && event.request.method !== 'HEAD') {
			init.body = await event.request.arrayBuffer();
		}

		try {
			const res = await fetch(targetUrl, init);
			return res;
		} catch (err) {
			return new Response(JSON.stringify({
				error: {
					code: 'GATEWAY_ERROR',
					message: 'Failed to proxy request to backend service: ' + err.message
				}
			}), {
				status: 502,
				headers: { 'Content-Type': 'application/json' }
			});
		}
	}

	return resolve(event);
}
