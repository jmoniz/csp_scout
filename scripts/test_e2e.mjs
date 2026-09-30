/**
 * End-to-end integration test verifying real-world CSP report collection,
 * origin normalization, wildcard deduction, triage transitions, and policy generation.
 */

const BACKEND_URL = process.env.TEST_BACKEND_URL || 'http://localhost:8080';

async function run() {
	console.log(`\n🚀 Starting End-to-End Verification against ${BACKEND_URL}...\n`);

	// 1. Healthcheck
	const healthRes = await fetch(`${BACKEND_URL}/health`);
	if (!healthRes.ok) throw new Error(`Healthcheck failed: ${healthRes.statusText}`);
	const health = await healthRes.json();
	console.log(`✅ [1/7] Healthcheck status: ${health.status}`);

	// 2. Create Session
	const sessionRes = await fetch(`${BACKEND_URL}/api/sessions`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			name: 'Customer Portal SPA E2E Test',
			target_origin: 'https://app.example.com',
			description: 'Automated end-to-end test session'
		})
	});
	if (sessionRes.status !== 201) throw new Error(`Failed to create session: ${sessionRes.statusText}`);
	const session = await sessionRes.json();
	const sessionId = session.id;
	console.log(`✅ [2/7] Session created: ID=${sessionId}, Report-URI=${session.report_uri}`);

	// 3. Send CSP violation report
	const reportPayload = {
		'csp-report': {
			'document-uri': 'https://app.example.com/dashboard',
			'referrer': 'https://app.example.com/',
			'violated-directive': 'connect-src',
			'effective-directive': 'connect-src',
			'original-policy': "default-src 'self'",
			'disposition': 'report',
			'blocked-uri': 'https://api.example.com/v1/users',
			'line-number': 128,
			'column-number': 45,
			'source-file': 'https://app.example.com/bundle.js',
			'status-code': 200
		}
	};

	const ingestRes = await fetch(`${BACKEND_URL}/api/reports/${sessionId}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/csp-report' },
		body: JSON.stringify(reportPayload)
	});
	if (ingestRes.status !== 204) throw new Error(`Failed to ingest report: ${ingestRes.status}`);

	// Send duplicate report to test counter increment
	await fetch(`${BACKEND_URL}/api/reports/${sessionId}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/csp-report' },
		body: JSON.stringify(reportPayload)
	});
	console.log(`✅ [3/7] Ingested 2 identical CSP violation reports`);

	// 4. Verify Triage & Aggregation
	const violationsRes = await fetch(`${BACKEND_URL}/api/sessions/${sessionId}/violations?directive=connect-src`);
	const violations = await violationsRes.json();
	if (violations.length !== 1) throw new Error(`Expected 1 violation group, got ${violations.length}`);
	const v = violations[0];

	if (v.origin_host !== 'https://api.example.com') throw new Error(`Unexpected origin_host: ${v.origin_host}`);
	if (v.suggested_wildcard !== 'https://*.example.com') throw new Error(`Unexpected wildcard: ${v.suggested_wildcard}`);
	if (v.count !== 2) throw new Error(`Expected count 2, got ${v.count}`);
	if (v.status !== 'pending') throw new Error(`Expected pending status, got ${v.status}`);
	console.log(`✅ [4/7] Verified aggregated violation group: ${v.origin_host} (Count: ${v.count}, Wildcard: ${v.suggested_wildcard})`);

	// 5. Test Violation Samples
	const samplesRes = await fetch(`${BACKEND_URL}/api/sessions/${sessionId}/violations/${v.id}/samples`);
	const samples = await samplesRes.json();
	if (samples.length !== 2) throw new Error(`Expected 2 samples, got ${samples.length}`);
	console.log(`✅ [5/8] Verified evidence samples: ${samples[0].blocked_uri} at line ${samples[0].line_number}`);

	// 6. Test 'self' Detection & Bulk Approval
	const selfReport = {
		'csp-report': {
			'document-uri': 'https://app.example.com/dashboard',
			'blocked-uri': '/static/js/vendor.chunk.js',
			'effective-directive': 'script-src',
			'line-number': 1
		}
	};
	const selfIngestRes = await fetch(`${BACKEND_URL}/api/reports/${sessionId}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/csp-report' },
		body: JSON.stringify(selfReport)
	});
	if (selfIngestRes.status !== 204) throw new Error('Failed to ingest self report');

	const selfViolationsRes = await fetch(`${BACKEND_URL}/api/sessions/${sessionId}/violations?source=self`);
	const selfViolations = await selfViolationsRes.json();
	if (selfViolations.length !== 1 || !selfViolations[0].is_self) {
		throw new Error(`Expected 1 'self' violation, got ${selfViolations.length}`);
	}

	const approveSelfRes = await fetch(`${BACKEND_URL}/api/sessions/${sessionId}/violations/approve-self`, {
		method: 'POST'
	});
	if (!approveSelfRes.ok) throw new Error('Failed to bulk approve self violations');
	const approveSelfData = await approveSelfRes.json();
	if (approveSelfData.approved !== 1) throw new Error(`Expected 1 self approved, got ${approveSelfData.approved}`);
	console.log(`✅ [6/8] Ingested relative path, verified is_self=true, and executed BulkApproveSelf`);

	// 7. Test Approval & Dynamic Policy Generation
	// A) Approve Wildcard
	const patchRes = await fetch(`${BACKEND_URL}/api/sessions/${sessionId}/violations/${v.id}`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ status: 'approved_wildcard' })
	});
	if (!patchRes.ok) throw new Error('Failed to patch violation status');

	const policyRes = await fetch(`${BACKEND_URL}/api/sessions/${sessionId}/policy`);
	const policy = await policyRes.json();
	if (!policy.raw.includes('connect-src https://*.example.com') || !policy.raw.includes("script-src 'self'")) {
		throw new Error(`Policy does not contain expected rules: ${policy.raw}`);
	}
	console.log(`✅ [7/8] Verified CSP compilation with wildcard & 'self':\n     Raw: "${policy.raw}"\n     Nginx: ${policy.nginx}`);

	// 8. Cleanup session
	const delRes = await fetch(`${BACKEND_URL}/api/sessions/${sessionId}`, { method: 'DELETE' });
	if (delRes.status !== 204) throw new Error('Failed to delete test session');
	console.log(`✅ [8/8] Cleaned up test session (${sessionId})\n`);

	console.log('🎉 All End-to-End verification checks passed successfully!\n');
}

run().catch((err) => {
	console.error('❌ E2E Verification failed:', err);
	process.exit(1);
});
