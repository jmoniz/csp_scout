<script>
	import { goto } from '$app/navigation';
	import { toast } from '$lib/toast.js';

	let { open = $bindable(false), session = null } = $props();

	let activeGuideTab = $state('lazyheader'); // 'lazyheader' | 'native'

	function downloadProfileJson() {
		if (!session) return;
		const headerVal = session.header_snippet?.replace(/^Content-Security-Policy-Report-Only:\s*/, '') || `default-src 'none'; form-action 'none'; frame-ancestors 'none'; report-uri ${session.report_uri};`;
		const filterOrigin = session.target_origin ? (session.target_origin.replace(/\/+$/, '') + '/*') : '*';
		const profile = {
			id: `profile_cspscout_${Date.now()}`,
			name: `CSP Scout (${session.target_origin ? session.target_origin.replace(/^https?:\/\//, '') : 'Session'})`,
			color: '#6366f1',
			enabled: true,
			urlFilter: filterOrigin,
			requestHeaders: [],
			responseHeaders: [
				{
					id: `h_${Date.now()}`,
					name: 'Content-Security-Policy-Report-Only',
					value: headerVal,
					operation: 'set',
					enabled: true
				}
			],
			redirects: []
		};
		const blob = new Blob([JSON.stringify(profile, null, 2)], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		const safeName = (session.target_origin || 'session').replace(/[^a-z0-9_-]+/gi, '_');
		a.download = `cspscout_${safeName}.headereditor.json`;
		a.click();
		URL.revokeObjectURL(url);
		toast.success('Downloaded profile for Lazy Header Editor!');
	}

	function copyToClipboard(text, label) {
		if (!text) return;
		navigator.clipboard.writeText(text).then(
			() => toast.success(`Copied ${label} to clipboard!`),
			() => toast.error(`Failed to copy ${label}`)
		);
	}

	function close() {
		open = false;
	}

	function goToSession() {
		if (session && session.id) {
			open = false;
			goto(`/session/${session.id}`);
		}
	}

	function handleKeydown(e) {
		if (e.key === 'Escape' && open) {
			close();
		}
	}

	function trapFocus(node) {
		const focusableElements = node.querySelectorAll(
			'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
		);
		const first = focusableElements[0];
		const last = focusableElements[focusableElements.length - 1];

		setTimeout(() => {
			first?.focus();
		}, 20);

		function onKeyDown(e) {
			if (e.key !== 'Tab') return;
			if (e.shiftKey) {
				if (document.activeElement === first) {
					e.preventDefault();
					last?.focus();
				}
			} else {
				if (document.activeElement === last) {
					e.preventDefault();
					first?.focus();
				}
			}
		}

		node.addEventListener('keydown', onKeyDown);
		return {
			destroy() {
				node.removeEventListener('keydown', onKeyDown);
			}
		};
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open && session}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-fade-in"
		role="dialog"
		aria-modal="true"
		aria-labelledby="header-setup-modal-title"
		tabindex="-1"
		onclick={(e) => { if (e.target === e.currentTarget) close(); }}
		onkeydown={(e) => { if (e.key === 'Escape') close(); }}
	>
		<div
			use:trapFocus
			class="bg-slate-900 border border-slate-800 rounded-2xl max-w-2xl w-full p-6 shadow-2xl relative max-h-[90vh] overflow-y-auto"
		>
			<!-- Close button -->
			<button
				onclick={close}
				class="absolute top-5 right-5 text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition-colors"
				aria-label="Close"
			>
				<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>

			<div class="flex items-center gap-3 mb-4">
				<div class="w-10 h-10 rounded-xl bg-indigo-600/20 text-indigo-400 border border-indigo-500/30 flex items-center justify-center shrink-0">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
					</svg>
				</div>
				<div>
					<h3 id="header-setup-modal-title" class="text-lg font-bold text-white">Browser Header Setup</h3>
					<p class="text-xs text-slate-400">Configure your browser extension or proxy to inject the Report-Only header.</p>
				</div>
			</div>

			<!-- Direct values to copy -->
			<div class="space-y-3 mb-6 bg-slate-950/60 p-4 rounded-xl border border-slate-800">
				<div>
					<div class="flex justify-between items-center mb-1">
						<span class="text-xs font-medium text-slate-400">Header Name</span>
						<button
							onclick={() => copyToClipboard('Content-Security-Policy-Report-Only', 'Header Name')}
							class="text-[11px] text-indigo-400 hover:text-indigo-300 font-medium"
						>
							Copy Name
						</button>
					</div>
					<code class="block text-xs bg-slate-900 px-3 py-2 rounded border border-slate-800 text-amber-300 font-mono select-all">
						Content-Security-Policy-Report-Only
					</code>
				</div>

				<div>
					<div class="flex justify-between items-center mb-1">
						<span class="text-xs font-medium text-slate-400">Header Value (Discovery Mode)</span>
						<button
							onclick={() => copyToClipboard(session.header_snippet?.replace(/^Content-Security-Policy-Report-Only:\s*/, '') || `default-src 'none'; form-action 'none'; frame-ancestors 'none'; report-uri ${session.report_uri};`, 'Header Value')}
							class="text-[11px] text-indigo-400 hover:text-indigo-300 font-medium"
						>
							Copy Value
						</button>
					</div>
					<code class="block text-xs bg-slate-900 px-3 py-2 rounded border border-slate-800 text-emerald-300 font-mono break-all select-all">
						{session.header_snippet?.replace(/^Content-Security-Policy-Report-Only:\s*/, '') || `default-src 'none'; form-action 'none'; frame-ancestors 'none'; report-uri ${session.report_uri};`}
					</code>
				</div>

				<div>
					<div class="flex justify-between items-center mb-1">
						<span class="text-xs font-medium text-slate-400">Report-URI Endpoint</span>
						<button
							onclick={() => copyToClipboard(session.report_uri, 'Report-URI')}
							class="text-[11px] text-indigo-400 hover:text-indigo-300 font-medium"
						>
							Copy URL
						</button>
					</div>
					<code class="block text-xs bg-slate-900 px-3 py-2 rounded border border-slate-800 text-indigo-300 font-mono break-all select-all">
						{session.report_uri}
					</code>
				</div>
			</div>

			<!-- Extension Guides Tabs -->
			<div class="border-b border-slate-800 flex gap-4 text-xs font-medium mb-4">
				<button
					onclick={() => activeGuideTab = 'lazyheader'}
					class="pb-2 border-b-2 transition-colors {activeGuideTab === 'lazyheader' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-slate-400 hover:text-slate-300'}"
				>
					Lazy Header Editor (Recommended)
				</button>
				<button
					onclick={() => activeGuideTab = 'native'}
					class="pb-2 border-b-2 transition-colors {activeGuideTab === 'native' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-slate-400 hover:text-slate-300'}"
				>
					Server Injection
				</button>
			</div>

			{#if activeGuideTab === 'lazyheader'}
				<div class="text-xs text-slate-300 space-y-3.5">
					<div class="bg-indigo-950/30 border border-indigo-900/50 rounded-lg p-3 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
						<div>
							<div class="font-semibold text-indigo-300 flex items-center gap-1.5">
								<span>Lazy Header Editor</span>
								<span class="text-[10px] bg-emerald-950 text-emerald-300 border border-emerald-800/60 px-1.5 py-0.5 rounded font-medium">Free &bull; No Ads &bull; MV3</span>
							</div>
							<p class="text-[11px] text-slate-400 mt-0.5">
								Ultra-lightweight (17 KB), completely local, zero telemetry.
							</p>
						</div>
						<div class="flex items-center gap-2 shrink-0">
							<a
								href="https://chromewebstore.google.com/detail/lazy-header-editor-modify/pcmlilcdikmaadfkkjejdghnikdibakg"
								target="_blank"
								rel="noopener noreferrer"
								class="px-2.5 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-md text-[11px] font-medium transition-colors inline-flex items-center gap-1"
							>
								<span>Get Extension</span>
								<svg class="w-3 h-3 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
								</svg>
							</a>
							<button
								onclick={downloadProfileJson}
								class="px-2.5 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-md text-[11px] font-medium transition-colors shadow-sm inline-flex items-center gap-1"
							>
								<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
								</svg>
								<span>Download Profile (.json)</span>
							</button>
						</div>
					</div>

					<div class="space-y-2">
						<h4 class="font-semibold text-slate-200 text-xs">Setup Options:</h4>
						<div class="grid grid-cols-1 md:grid-cols-2 gap-3">
							<div class="bg-slate-950/40 p-3 rounded-lg border border-slate-800/80">
								<span class="text-indigo-400 font-semibold text-[11px] block mb-1">Option A: 1-Click Profile Import (Fastest)</span>
								<ol class="list-decimal list-inside space-y-1 text-[11px] text-slate-300">
									<li>Install <a href="https://chromewebstore.google.com/detail/lazy-header-editor-modify/pcmlilcdikmaadfkkjejdghnikdibakg" target="_blank" rel="noopener noreferrer" class="text-indigo-400 underline hover:text-indigo-300">Lazy Header Editor</a>.</li>
									<li>Click <strong>Download Profile (.json)</strong> above.</li>
									<li>Click the extension icon &rarr; <strong>⋮</strong> (Menu) &rarr; <strong>Import profile</strong> and select the file.</li>
									<li>Turn the top-right master switch <strong>ON</strong> and reload your web app!</li>
								</ol>
							</div>

							<div class="bg-slate-950/40 p-3 rounded-lg border border-slate-800/80">
								<span class="text-indigo-400 font-semibold text-[11px] block mb-1">Option B: Manual Configuration</span>
								<ol class="list-decimal list-inside space-y-1 text-[11px] text-slate-300">
									<li>Open extension &rarr; <strong>Response Headers</strong> tab &rarr; <strong>+ Add header</strong>.</li>
									<li>Name: <code class="bg-slate-800 text-amber-300 px-1 py-0.5 rounded">Content-Security-Policy-Report-Only</code></li>
									<li>Value: Copy <strong>Header Value</strong> from above (Operation: <code class="bg-slate-800 text-slate-300 px-1 py-0.5 rounded">Set</code>).</li>
									<li>In <strong>Filter</strong> tab, set URL: <code class="bg-slate-800 text-indigo-300 px-1 py-0.5 rounded">{session.target_origin ? `${session.target_origin}/*` : '*'}</code></li>
									<li>Turn switch <strong>ON</strong> and reload your target app.</li>
								</ol>
							</div>
						</div>
					</div>
				</div>
			{:else}
				<div class="text-xs text-slate-300 space-y-2">
					<p class="text-slate-400">If you control the web server or reverse proxy, add the header in your server configuration:</p>
					<div class="bg-slate-950 p-3 rounded-lg border border-slate-800 font-mono text-[11px] text-slate-300 space-y-2">
						<p class="text-slate-500 font-sans text-xs"># Nginx config</p>
						<p>add_header Content-Security-Policy-Report-Only "default-src 'none'; form-action 'none'; frame-ancestors 'none'; report-uri {session.report_uri};";</p>
					</div>
				</div>
			{/if}

			<div class="mt-6 flex flex-col-reverse sm:flex-row sm:items-center justify-between gap-3 pt-4 border-t border-slate-800">
				<button
					onclick={close}
					class="px-3.5 py-2 text-slate-400 hover:text-slate-200 text-xs font-medium rounded-lg hover:bg-slate-800/80 transition-colors text-center"
				>
					Stay on Dashboard
				</button>
				<button
					onclick={goToSession}
					class="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold shadow-md shadow-indigo-600/20 transition-all flex items-center justify-center gap-1.5"
				>
					<span>Enter Triage Workspace</span>
					<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
					</svg>
				</button>
			</div>
		</div>
	</div>
{/if}
