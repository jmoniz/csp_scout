import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';

const BASE_URL = 'http://localhost:3000';

async function main() {
	console.log('1. Creating showcase session...');
	const sessionRes = await fetch(`${BASE_URL}/api/sessions`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			name: 'Customer Portal SPA',
			target_origin: 'https://app.example.com',
			description: 'Production CSP policy generation and triage'
		})
	});

	if (!sessionRes.ok) {
		throw new Error(`Failed to create session: ${sessionRes.statusText}`);
	}

	const session = await sessionRes.json();
	const sid = session.id;
	console.log(`Session created: ${sid}`);

	// Helper to send report
	async function sendReport(directive, blockedUri, count = 1) {
		const payload = {
			'csp-report': {
				'document-uri': 'https://app.example.com/dashboard',
				'referrer': 'https://app.example.com/',
				'violated-directive': directive,
				'effective-directive': directive,
				'original-policy': "default-src 'none'",
				'disposition': 'report',
				'blocked-uri': blockedUri,
				'line-number': Math.floor(Math.random() * 200) + 1,
				'column-number': Math.floor(Math.random() * 60) + 1,
				'source-file': 'https://app.example.com/assets/app.js',
				'status-code': 200
			}
		};

		for (let i = 0; i < count; i++) {
			await fetch(`${BASE_URL}/api/reports/${sid}`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/csp-report' },
				body: JSON.stringify(payload)
			});
		}
	}

	console.log('2. Ingesting realistic CSP violation reports...');
	await sendReport('script-src', 'https://static.cloudflareinsights.com/beacon.min.js', 12);
	await sendReport('script-src', 'https://www.googletagmanager.com/gtm.js?id=GTM-5K89', 8);
	await sendReport('script-src', 'https://js.stripe.com/v3/', 6);
	await sendReport('script-src', "'unsafe-inline'", 3);

	await sendReport('connect-src', 'https://app.example.com/api/v1/graphql', 45);
	await sendReport('connect-src', 'https://api.stripe.com/v1/tokens', 14);
	await sendReport('connect-src', 'https://region1.google-analytics.com/g/collect', 22);
	await sendReport('connect-src', 'https://o12345.ingest.sentry.io/api/123/envelope/', 5);

	await sendReport('img-src', 'https://images.unsplash.com/photo-1542291026-7eec264c27ff', 31);
	await sendReport('img-src', 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0i...', 9);

	await sendReport('style-src', 'https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700', 16);
	await sendReport('font-src', 'https://fonts.gstatic.com/s/inter/v13/UcCO3FwrK3iLTeHuS_fvQtMwCp50KnMw2boKoduKmMEVuLyfAZ9hiA.woff2', 16);

	console.log('3. Triaging violation groups...');
	const violRes = await fetch(`${BASE_URL}/api/sessions/${sid}/violations`);
	const violations = await violRes.json();

	async function updateStatus(originSubstr, status) {
		const target = violations.find(v => v.origin_host.includes(originSubstr) || (v.suggested_wildcard && v.suggested_wildcard.includes(originSubstr)));
		if (target) {
			await fetch(`${BASE_URL}/api/sessions/${sid}/violations/${target.id}`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status })
			});
			console.log(`Updated ${target.origin_host} to ${status}`);
		}
	}

	await updateStatus('app.example.com', 'approved_self');
	await updateStatus('stripe.com', 'approved_origin');
	await updateStatus('cloudflareinsights.com', 'approved_origin');
	await updateStatus('googletagmanager.com', 'approved_wildcard');
	await updateStatus('googleapis.com', 'approved_origin');
	await updateStatus('gstatic.com', 'approved_wildcard');
	await updateStatus('unsplash.com', 'approved_origin');
	await updateStatus('data:', 'approved_origin');
	await updateStatus('unsafe-inline', 'rejected');

	// Ensure output directory exists
	const outDir = path.resolve('docs/images');
	if (!fs.existsSync(outDir)) {
		fs.mkdirSync(outDir, { recursive: true });
	}

	const screenshotPath = path.join(outDir, 'dashboard.png');
	const sessionUrl = `${BASE_URL}/session/${sid}`;

	console.log(`4. Capturing screenshot of ${sessionUrl}...`);
	const chromePath = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
	const args = [
		'--headless=new',
		`--screenshot=${screenshotPath}`,
		'--window-size=1440,940',
		'--virtual-time-budget=4000',
		'--hide-scrollbars',
		sessionUrl
	];

	const { execFileSync } = await import('child_process');
	execFileSync(chromePath, args, { stdio: 'inherit' });
	console.log(`✅ Screenshot saved to ${screenshotPath}`);

	return sid;
}

main().catch(err => {
	console.error('Error:', err);
	process.exit(1);
});
