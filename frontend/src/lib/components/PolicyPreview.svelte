<script>
	import { api } from '$lib/api.js';
	import { toast } from '$lib/toast.js';

	let { sessionId = '', initialSettings = null, onSettingsChanged = () => {} } = $props();

	let policyData = $state(null);
	const defaultSettings = {
		default_src: "'none'",
		form_action: "'none'",
		frame_ancestors: "'none'",
		upgrade_insecure_requests: false,
		block_all_mixed_content: false,
		report_only: false,
		custom_directives: ''
	};
	let settings = $state({ ...defaultSettings });

	let activeFormat = $state('header'); // 'header' | 'raw' | 'meta' | 'nginx' | 'apache'
	let showSettingsModal = $state(false);
	let savingSettings = $state(false);
	let loadingPolicy = $state(false);

	export async function loadPolicy() {
		if (!sessionId) return;
		loadingPolicy = true;
		try {
			policyData = await api.getPolicy(sessionId);
		} catch (err) {
			toast.error('Failed to compile CSP policy');
		} finally {
			loadingPolicy = false;
		}
	}

	$effect(() => {
		if (sessionId) {
			loadPolicy();
		}
	});

	$effect(() => {
		if (initialSettings) {
			settings = { ...initialSettings };
		}
	});

	async function toggleReportOnly() {
		const newReportOnly = !settings.report_only;
		settings.report_only = newReportOnly;
		try {
			await api.updateSettings(sessionId, settings);
			toast.success(newReportOnly ? 'Switched to Report-Only mode' : 'Switched to Enforce mode');
			await loadPolicy();
			onSettingsChanged();
		} catch (err) {
			settings.report_only = !newReportOnly;
			toast.error('Failed to update mode');
		}
	}

	async function saveSettings(e) {
		e.preventDefault();
		savingSettings = true;
		try {
			const updated = await api.updateSettings(sessionId, settings);
			settings = updated;
			toast.success('Policy configuration saved!');
			showSettingsModal = false;
			await loadPolicy();
			onSettingsChanged();
		} catch (err) {
			toast.error('Failed to save settings: ' + err.message);
		} finally {
			savingSettings = false;
		}
	}

	function getCurrentContent() {
		if (!policyData) return '';
		switch (activeFormat) {
			case 'header': return policyData.header || '';
			case 'raw': return policyData.raw || '';
			case 'meta': return policyData.meta || '';
			case 'nginx': return policyData.nginx || '';
			case 'apache': return policyData.apache || '';
			default: return policyData.raw || '';
		}
	}

	function copyActiveFormat() {
		const content = getCurrentContent();
		if (!content) return;
		navigator.clipboard.writeText(content).then(
			() => toast.success(`Copied ${activeFormat.toUpperCase()} snippet!`),
			() => toast.error('Failed to copy to clipboard')
		);
	}
</script>

<div class="bg-slate-900/80 border border-slate-800 rounded-2xl p-5 space-y-4 shadow-xl">
	<!-- Top Bar: Mode Toggle & Settings -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800 pb-4">
		<div class="flex items-center gap-2">
			<div class="w-8 h-8 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center justify-center shrink-0">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
				</svg>
			</div>
			<div>
				<h3 class="text-sm font-bold text-white">Live Compiled CSP</h3>
				<p class="text-[11px] text-slate-400">Updates dynamically as you approve or reject violation groups.</p>
			</div>
		</div>

		<div class="flex items-center gap-2">
			<!-- Mode Toggle Button -->
			<button
				onclick={toggleReportOnly}
				class="px-2.5 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1.5 transition-all {settings.report_only ? 'bg-amber-950/80 text-amber-300 border border-amber-700/80' : 'bg-emerald-950/80 text-emerald-300 border border-emerald-700/80'}"
				title="Toggle between Enforce and Report-Only modes"
			>
				<span class="w-2 h-2 rounded-full {settings.report_only ? 'bg-amber-400' : 'bg-emerald-400'}"></span>
				<span>{settings.report_only ? 'Report-Only Mode' : 'Enforce Mode'}</span>
			</button>

			<!-- Settings Button -->
			<button
				onclick={() => showSettingsModal = true}
				class="p-1.5 text-slate-400 hover:text-white bg-slate-800 hover:bg-slate-700 rounded-lg transition-colors border border-slate-700"
				title="Configure base policy directives"
			>
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
				</svg>
			</button>
		</div>
	</div>

	<!-- Format Selectors & Copy Button -->
	<div class="flex items-center justify-between gap-2">
		<div class="flex items-center gap-1 bg-slate-950 p-1 rounded-lg border border-slate-800 text-[11px] font-medium overflow-x-auto">
			<button
				onclick={() => activeFormat = 'header'}
				class="px-2.5 py-1 rounded transition-colors {activeFormat === 'header' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'}"
			>
				HTTP Header
			</button>
			<button
				onclick={() => activeFormat = 'raw'}
				class="px-2.5 py-1 rounded transition-colors {activeFormat === 'raw' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'}"
			>
				Raw Directives
			</button>
			<button
				onclick={() => activeFormat = 'meta'}
				class="px-2.5 py-1 rounded transition-colors {activeFormat === 'meta' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'}"
			>
				HTML &lt;meta&gt;
			</button>
			<button
				onclick={() => activeFormat = 'nginx'}
				class="px-2.5 py-1 rounded transition-colors {activeFormat === 'nginx' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'}"
			>
				Nginx
			</button>
			<button
				onclick={() => activeFormat = 'apache'}
				class="px-2.5 py-1 rounded transition-colors {activeFormat === 'apache' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'}"
			>
				Apache
			</button>
		</div>

		<button
			onclick={copyActiveFormat}
			class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold shadow transition-all flex items-center gap-1.5 shrink-0"
		>
			<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
			</svg>
			<span>Copy</span>
		</button>
	</div>

	<!-- Output Code Window -->
	<div class="relative bg-slate-950 border border-slate-800 rounded-xl p-4 font-mono text-xs overflow-x-auto min-h-[100px] flex items-center">
		{#if loadingPolicy}
			<div class="w-full text-center text-slate-500 py-4">Compiling policy...</div>
		{:else if policyData}
			<pre class="text-emerald-300 whitespace-pre-wrap break-all select-all leading-relaxed">{getCurrentContent()}</pre>
		{:else}
			<div class="w-full text-center text-slate-500 py-4">No policy compiled yet.</div>
		{/if}
	</div>
</div>

<!-- Settings Modal Dialog -->
{#if showSettingsModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-fade-in" role="dialog" aria-modal="true">
		<div class="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl relative">
			<button
				onclick={() => showSettingsModal = false}
				class="absolute top-5 right-5 text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition-colors"
				aria-label="Close"
			>
				<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>

			<h3 class="text-base font-bold text-white mb-1">Base Policy Directives</h3>
			<p class="text-xs text-slate-400 mb-4">Fine-tune fallback and framing directives for this session.</p>

			<form onsubmit={saveSettings} class="space-y-4">
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
					<div>
						<label for="default-src" class="block text-xs font-semibold text-slate-300 mb-1">default-src</label>
						<input
							id="default-src"
							type="text"
							bind:value={settings.default_src}
							class="w-full bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:border-indigo-500"
							placeholder="'none'"
						/>
					</div>

					<div>
						<label for="form-action" class="block text-xs font-semibold text-slate-300 mb-1">form-action</label>
						<input
							id="form-action"
							type="text"
							bind:value={settings.form_action}
							class="w-full bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:border-indigo-500"
							placeholder="'none'"
						/>
					</div>

					<div>
						<label for="frame-ancestors" class="block text-xs font-semibold text-slate-300 mb-1">frame-ancestors</label>
						<input
							id="frame-ancestors"
							type="text"
							bind:value={settings.frame_ancestors}
							class="w-full bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:border-indigo-500"
							placeholder="'none'"
						/>
					</div>
				</div>

				<div class="space-y-2 pt-2 border-t border-slate-800">
					<label class="flex items-center gap-2 cursor-pointer text-xs text-slate-300">
						<input
							type="checkbox"
							bind:checked={settings.upgrade_insecure_requests}
							class="rounded border-slate-700 bg-slate-950 text-indigo-600 focus:ring-indigo-500"
						/>
						<span>Include <code class="text-indigo-400 font-mono">upgrade-insecure-requests</code></span>
					</label>

					<label class="flex items-center gap-2 cursor-pointer text-xs text-slate-300">
						<input
							type="checkbox"
							bind:checked={settings.block_all_mixed_content}
							class="rounded border-slate-700 bg-slate-950 text-indigo-600 focus:ring-indigo-500"
						/>
						<span>Include <code class="text-indigo-400 font-mono">block-all-mixed-content</code></span>
					</label>
				</div>

				<div>
					<label for="custom-directives" class="block text-xs font-semibold text-slate-300 mb-1">Custom Extra Directives</label>
					<textarea
						id="custom-directives"
						bind:value={settings.custom_directives}
						rows="2"
						placeholder="e.g. sandbox allow-scripts; base-uri 'self'"
						class="w-full bg-slate-950 border border-slate-800 rounded-lg px-2.5 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:border-indigo-500"
					></textarea>
				</div>

				<div class="pt-3 flex items-center justify-end gap-2 border-t border-slate-800">
					<button
						type="button"
						onclick={() => showSettingsModal = false}
						class="px-3 py-1.5 text-xs font-medium text-slate-400 hover:text-white rounded-lg hover:bg-slate-800 transition-colors"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={savingSettings}
						class="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold shadow transition-colors"
					>
						{savingSettings ? 'Saving...' : 'Save Settings'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
