<script>
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib/api.js';
	import { toast } from '$lib/toast.js';
	import ViolationTable from '$lib/components/ViolationTable.svelte';
	import PolicyPreview from '$lib/components/PolicyPreview.svelte';
	import HeaderModal from '$lib/components/HeaderModal.svelte';

	let sessionId = $derived(page.params.id);

	let sessionData = $state(null);
	let loading = $state(true);
	let error = $state('');

	let showHeaderModal = $state(false);

	let violationTableRef = $state();
	let policyPreviewRef = $state();

	// Live Capture Reactive Polling
	let liveMode = $state(true);
	let newActivity = $state(false);
	let isPolling = false;
	const POLL_INTERVAL_MS = 3000;
	let pollTimer = null;
	let activityResetTimer = null;

	async function loadSession() {
		if (!sessionId) return;
		loading = true;
		error = '';
		try {
			sessionData = await api.getSession(sessionId);
		} catch (err) {
			error = err.message || 'Failed to load session details';
			toast.error(error);
		} finally {
			loading = false;
		}
	}

	async function pollMetrics() {
		if (!sessionId || !liveMode || isPolling) return;
		if (typeof document !== 'undefined' && document.hidden) return;

		isPolling = true;
		try {
			const fresh = await api.getSession(sessionId);
			if (!sessionData) {
				sessionData = fresh;
				return;
			}

			const reportsChanged = fresh.total_reports !== sessionData.total_reports;
			const pendingChanged = fresh.pending_violations !== sessionData.pending_violations;
			const selfChanged = fresh.self_violations !== sessionData.self_violations;
			const approvedChanged = fresh.approved_rules !== sessionData.approved_rules;

			if (reportsChanged || pendingChanged || selfChanged) {
				sessionData = fresh;
				newActivity = true;
				if (activityResetTimer) clearTimeout(activityResetTimer);
				activityResetTimer = setTimeout(() => {
					newActivity = false;
				}, 1200);

				// Silently refresh violation table without showing full screen loader
				if (violationTableRef && violationTableRef.reload) {
					violationTableRef.reload(true);
				}
			} else {
				sessionData = fresh;
			}

			if (approvedChanged && policyPreviewRef && policyPreviewRef.loadPolicy) {
				policyPreviewRef.loadPolicy();
			}
		} catch {
			// Silently ignore background polling errors (e.g. transient network hiccup)
		} finally {
			isPolling = false;
		}
	}

	function startPolling() {
		stopPolling();
		pollTimer = setInterval(pollMetrics, POLL_INTERVAL_MS);
	}

	function stopPolling() {
		if (pollTimer) {
			clearInterval(pollTimer);
			pollTimer = null;
		}
	}

	function toggleLiveMode() {
		liveMode = !liveMode;
		if (liveMode) {
			pollMetrics();
			startPolling();
		} else {
			stopPolling();
		}
	}

	function handleVisibilityChange() {
		if (typeof document !== 'undefined' && !document.hidden && liveMode) {
			pollMetrics();
		}
	}

	onMount(() => {
		loadSession();
		startPolling();
		if (typeof document !== 'undefined') {
			document.addEventListener('visibilitychange', handleVisibilityChange);
		}
	});

	onDestroy(() => {
		stopPolling();
		if (activityResetTimer) clearTimeout(activityResetTimer);
		if (typeof document !== 'undefined') {
			document.removeEventListener('visibilitychange', handleVisibilityChange);
		}
	});

	async function handleViolationUpdated() {
		// Refresh session summary metrics and the live policy preview
		try {
			sessionData = await api.getSession(sessionId);
			if (policyPreviewRef && policyPreviewRef.loadPolicy) {
				policyPreviewRef.loadPolicy();
			}
		} catch {
			// silently ignore minor metric refresh error
		}
	}

	function handleSettingsChanged() {
		// When settings change in policy preview, refresh table or metrics if needed
		if (violationTableRef && violationTableRef.reload) {
			violationTableRef.reload();
		}
	}
</script>

<div class="space-y-6 animate-fade-in">
	<!-- Breadcrumb & Navigation -->
	<div class="flex items-center justify-between gap-4">
		<nav class="flex items-center gap-2 text-xs text-slate-400">
			<a href="/" class="hover:text-white transition-colors">Sessions</a>
			<span>/</span>
			<span class="text-slate-200 font-medium truncate max-w-[200px] sm:max-w-md">
				{sessionData ? sessionData.name : 'Loading...'}
			</span>
		</nav>

		<div class="flex items-center gap-2">
			<!-- Live Capture Mode Toggle -->
			<button
				type="button"
				onclick={toggleLiveMode}
				class="px-2.5 py-1.5 rounded-lg text-xs font-medium transition-all flex items-center gap-2 border {liveMode ? 'bg-emerald-950/60 hover:bg-emerald-900/60 text-emerald-300 border-emerald-700/60 shadow-sm' : 'bg-slate-800 hover:bg-slate-700 text-slate-400 border-slate-700'}"
				title={liveMode ? 'Live capture active (updating every 3s). Click to pause.' : 'Live capture paused. Click to resume.'}
			>
				{#if liveMode}
					<span class="relative flex h-2 w-2">
						<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
						<span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
					</span>
					<span>Live</span>
				{:else}
					<span class="inline-flex rounded-full h-2 w-2 bg-slate-500"></span>
					<span>Paused</span>
				{/if}
			</button>

			<button
				onclick={() => showHeaderModal = true}
				class="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 rounded-lg text-xs font-medium transition-colors flex items-center gap-1.5"
			>
				<svg class="w-3.5 h-3.5 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
				</svg>
				<span>Setup Header</span>
			</button>

			<button
				onclick={() => { loadSession(); if (violationTableRef) violationTableRef.reload(); if (policyPreviewRef) policyPreviewRef.loadPolicy(); }}
				class="p-1.5 text-slate-400 hover:text-white bg-slate-800 hover:bg-slate-700 rounded-lg transition-colors border border-slate-700"
				title="Manual reload"
			>
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
			</button>
		</div>
	</div>

	{#if loading}
		<div class="py-20 text-center text-slate-500 flex flex-col items-center justify-center gap-3">
			<svg class="animate-spin h-8 w-8 text-indigo-500" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
			</svg>
			<p class="text-xs">Loading session triage workspace...</p>
		</div>
	{:else if error}
		<div class="p-8 bg-rose-950/40 border border-rose-800/60 rounded-xl text-center max-w-lg mx-auto">
			<p class="text-rose-400 text-sm font-semibold mb-3">{error}</p>
			<a href="/" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold">
				Back to Sessions
			</a>
		</div>
	{:else if sessionData}
		<!-- Header Summary & Metrics -->
		<div class="bg-slate-900/60 border border-slate-800 rounded-2xl p-5 space-y-4">
			<div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
				<div>
					<h1 class="text-xl sm:text-2xl font-bold text-white tracking-tight">{sessionData.name}</h1>
					<div class="flex items-center gap-2 mt-1 font-mono text-xs text-indigo-400">
						<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
						</svg>
						<span class="select-all">{sessionData.target_origin}</span>
					</div>
					{#if sessionData.description}
						<p class="text-xs text-slate-400 mt-2 max-w-2xl">{sessionData.description}</p>
					{/if}
				</div>

				<div class="flex items-center gap-3 shrink-0 flex-wrap sm:flex-nowrap">
					<div class="bg-slate-950/80 px-3.5 py-2 rounded-xl border transition-all duration-300 text-center min-w-[85px] {newActivity ? 'border-indigo-500/80 bg-indigo-950/40 ring-2 ring-indigo-500/30' : 'border-slate-800'}">
						<div class="text-lg font-bold text-white flex items-center justify-center gap-1.5">
							<span>{sessionData.total_reports || 0}</span>
							{#if newActivity}
								<span class="inline-flex h-2 w-2 rounded-full bg-indigo-400 animate-ping"></span>
							{/if}
						</div>
						<div class="text-[10px] text-slate-400">Total Reports</div>
					</div>
					<div class="bg-slate-950/80 px-3.5 py-2 rounded-xl border transition-all duration-300 text-center min-w-[85px] {newActivity ? 'border-amber-500/80 bg-amber-950/40 ring-1 ring-amber-500/30' : 'border-slate-800'}">
						<div class="text-lg font-bold text-amber-400">{sessionData.pending_violations || 0}</div>
						<div class="text-[10px] text-slate-400">Pending Triage</div>
					</div>
					<div class="bg-slate-950/80 px-3.5 py-2 rounded-xl border transition-all duration-300 text-center min-w-[85px] {newActivity ? 'border-sky-500/80 bg-sky-950/40 ring-1 ring-sky-500/30' : 'border-slate-800'}">
						<div class="text-lg font-bold text-sky-400">{sessionData.self_violations || 0}</div>
						<div class="text-[10px] text-slate-400">'self' Matches</div>
					</div>
					<div class="bg-slate-950/80 px-3.5 py-2 rounded-xl border border-slate-800 text-center min-w-[85px]">
						<div class="text-lg font-bold text-emerald-400">{sessionData.approved_rules || 0}</div>
						<div class="text-[10px] text-slate-400">Approved Rules</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Main Workspace: Triage Table & Live CSP Preview -->
		<div class="space-y-6">
			<!-- Triage Section -->
			<section class="space-y-3">
				<div class="flex items-center justify-between">
					<h2 class="text-base font-bold text-white tracking-tight flex items-center gap-2">
						<span>Violations & Triage</span>
						<span class="text-xs font-normal text-slate-500">— Review blocked resources and allowlist origins</span>
					</h2>
				</div>

				<ViolationTable
					bind:this={violationTableRef}
					sessionId={sessionId}
					targetOrigin={sessionData.target_origin}
					onViolationUpdated={handleViolationUpdated}
				/>
			</section>

			<!-- Live Policy Preview & Export Drawer -->
			<section class="space-y-3 pt-4 border-t border-slate-800/80">
				<PolicyPreview
					bind:this={policyPreviewRef}
					sessionId={sessionId}
					initialSettings={sessionData.policy_setting}
					onSettingsChanged={handleSettingsChanged}
				/>
			</section>
		</div>
	{/if}
</div>

<!-- Extension Setup Guide Modal -->
<HeaderModal bind:open={showHeaderModal} session={sessionData} />
