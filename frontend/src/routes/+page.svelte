<script>
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib/api.js';
	import { toast } from '$lib/toast.js';
	import NewSessionModal from '$lib/components/NewSessionModal.svelte';
	import HeaderModal from '$lib/components/HeaderModal.svelte';

	let sessions = $state([]);
	let loading = $state(true);
	let error = $state('');

	let showNewModal = $state(false);
	let showHeaderModal = $state(false);
	let activeSessionForHeader = $state(null);

	let pollTimer = null;

	async function loadSessions() {
		loading = true;
		error = '';
		try {
			sessions = await api.getSessions();
		} catch (err) {
			error = err.message || 'Failed to load sessions';
			toast.error(error);
		} finally {
			loading = false;
		}
	}

	async function pollSessions() {
		if (typeof document !== 'undefined' && document.hidden) return;
		try {
			const fresh = await api.getSessions();
			sessions = fresh;
		} catch {
			// silently ignore quiet background polling error
		}
	}

	function handleSessionCreated(newSession) {
		sessions = [newSession, ...sessions];
		activeSessionForHeader = newSession;
		showHeaderModal = true;
	}

	let confirmingDeleteId = $state(null);
	let deleteTimeout = null;

	function promptDeleteSession(id) {
		confirmingDeleteId = id;
		if (deleteTimeout) clearTimeout(deleteTimeout);
		deleteTimeout = setTimeout(() => {
			if (confirmingDeleteId === id) {
				confirmingDeleteId = null;
			}
		}, 4000);
	}

	function cancelDeleteSession() {
		if (deleteTimeout) clearTimeout(deleteTimeout);
		confirmingDeleteId = null;
	}

	async function confirmDeleteSession(id, name) {
		if (deleteTimeout) clearTimeout(deleteTimeout);
		confirmingDeleteId = null;
		try {
			await api.deleteSession(id);
			sessions = sessions.filter((s) => s.id !== id);
			toast.success(`Session "${name}" deleted.`);
		} catch (err) {
			toast.error(err.message || 'Failed to delete session');
		}
	}

	function openHeaderGuide(sess) {
		activeSessionForHeader = sess;
		showHeaderModal = true;
	}

	function formatRelativeTime(dateStr) {
		if (!dateStr) return 'No reports yet';
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = now - date;
		const diffSec = Math.floor(diffMs / 1000);
		if (diffSec < 60) return 'Reported just now';
		const diffMin = Math.floor(diffSec / 60);
		if (diffMin < 60) return `Last report ${diffMin}m ago`;
		const diffHours = Math.floor(diffMin / 60);
		if (diffHours < 24) return `Last report ${diffHours}h ago`;
		const diffDays = Math.floor(diffHours / 24);
		return `Last report ${diffDays}d ago`;
	}

	function isRecentlyActive(session) {
		if (!session.total_reports || session.total_reports === 0) return false;
		if (!session.updated_at) return false;
		const diffMs = new Date() - new Date(session.updated_at);
		return diffMs < 1000 * 60 * 5; // active within last 5 minutes
	}

	function handleGlobalKeydown(e) {
		if ((e.key === 'n' || e.key === 'N') && !e.ctrlKey && !e.metaKey && !e.altKey) {
			const activeTag = document.activeElement?.tagName?.toLowerCase();
			if (activeTag === 'input' || activeTag === 'textarea' || activeTag === 'select') return;
			if (!showNewModal && !showHeaderModal) {
				e.preventDefault();
				showNewModal = true;
			}
		}
	}

	onMount(() => {
		loadSessions();
		pollTimer = setInterval(pollSessions, 4000);
	});

	onDestroy(() => {
		if (pollTimer) {
			clearInterval(pollTimer);
			pollTimer = null;
		}
		if (deleteTimeout) {
			clearTimeout(deleteTimeout);
			deleteTimeout = null;
		}
	});

	let totalReportsAll = $derived(sessions.reduce((acc, s) => acc + (s.total_reports || 0), 0));
	let totalPendingAll = $derived(sessions.reduce((acc, s) => acc + (s.pending_violations || 0), 0));
	let totalApprovedAll = $derived(sessions.reduce((acc, s) => acc + (s.approved_rules || 0), 0));
</script>

<svelte:window onkeydown={handleGlobalKeydown} />

<div class="space-y-8 animate-fade-in">
	<!-- Hero / Header Section -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-6">
		<div>
			<h1 class="text-2xl sm:text-3xl font-bold tracking-tight text-white">CSP Analysis Sessions</h1>
			<p class="text-sm text-slate-400 mt-1">
				Capture CSP violation reports in Report-Only mode, triage third-party resources, and craft clean policies.
			</p>
		</div>
		<div>
			<button
				onclick={() => showNewModal = true}
				class="inline-flex items-center gap-2 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl text-xs sm:text-sm font-semibold shadow-lg shadow-indigo-600/25 transition-all transform hover:-translate-y-0.5 active:translate-y-0 cursor-pointer"
			>
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				<span>New Session</span>
				<kbd class="hidden sm:inline-block font-mono text-[10px] bg-indigo-700/60 text-indigo-200 px-1.5 py-0.5 rounded border border-indigo-400/30 ml-0.5" title="Keyboard shortcut: Press 'N' to open">N</kbd>
			</button>
		</div>
	</div>

	<!-- Metric Highlights -->
	<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
		<div class="bg-slate-900/60 border border-slate-800/80 rounded-xl p-4 flex items-center gap-4">
			<div class="w-12 h-12 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 flex items-center justify-center shrink-0">
				<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
				</svg>
			</div>
			<div>
				<div class="text-2xl font-bold text-white tracking-tight">{totalReportsAll}</div>
				<div class="text-xs font-medium text-slate-400">Total Captured Reports</div>
			</div>
		</div>

		<div class="bg-slate-900/60 border border-slate-800/80 rounded-xl p-4 flex items-center gap-4">
			<div class="w-12 h-12 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20 flex items-center justify-center shrink-0">
				<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
				</svg>
			</div>
			<div>
				<div class="text-2xl font-bold text-white tracking-tight">{totalPendingAll}</div>
				<div class="text-xs font-medium text-slate-400">Pending Violations</div>
			</div>
		</div>

		<div class="bg-slate-900/60 border border-slate-800/80 rounded-xl p-4 flex items-center gap-4">
			<div class="w-12 h-12 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center justify-center shrink-0">
				<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
			</div>
			<div>
				<div class="text-2xl font-bold text-white tracking-tight">{totalApprovedAll}</div>
				<div class="text-xs font-medium text-slate-400">Approved Rules</div>
			</div>
		</div>
	</div>

	<!-- Sessions List -->
	{#if loading}
		<div class="py-16 text-center text-slate-500 flex flex-col items-center justify-center gap-3">
			<svg class="animate-spin h-8 w-8 text-indigo-500" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
			</svg>
			<p class="text-sm">Loading sessions...</p>
		</div>
	{:else if error}
		<div class="p-6 bg-rose-950/40 border border-rose-800/60 rounded-xl text-center">
			<p class="text-rose-400 text-sm font-semibold mb-2">{error}</p>
			<button
				onclick={loadSessions}
				class="px-3 py-1.5 bg-rose-800 hover:bg-rose-700 text-white rounded-lg text-xs font-medium"
			>
				Try Again
			</button>
		</div>
	{:else if sessions.length === 0}
		<!-- Empty state with onboarding guide -->
		<div class="bg-slate-900/40 border border-dashed border-slate-800 rounded-2xl p-10 text-center max-w-xl mx-auto space-y-4">
			<div class="w-14 h-14 rounded-2xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 flex items-center justify-center mx-auto">
				<svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
				</svg>
			</div>
			<div>
				<h3 class="text-base font-bold text-white">No active analysis sessions</h3>
				<p class="text-xs text-slate-400 mt-1 max-w-sm mx-auto">
					Create your first session to start capturing browser CSP reports in real time without blocking legitimate user traffic.
				</p>
			</div>
			<button
				onclick={() => showNewModal = true}
				class="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold shadow transition-colors"
			>
				Create New Session
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
			{#each sessions as session (session.id)}
				<div class="bg-slate-900/70 border border-slate-800 rounded-2xl p-5 hover:border-slate-700 transition-all flex flex-col justify-between group shadow-sm hover:shadow-md">
					<div class="space-y-3">
						<div class="flex items-start justify-between gap-2">
							<div class="space-y-1 min-w-0 flex-1">
								<div class="flex items-center gap-2">
									{#if isRecentlyActive(session)}
										<span class="relative flex h-2 w-2 shrink-0" title="Actively receiving CSP telemetry">
											<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
											<span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
										</span>
									{:else if session.total_reports > 0}
										<span class="inline-flex rounded-full h-2 w-2 bg-slate-500 shrink-0" title="Telemetry recorded"></span>
									{:else}
										<span class="inline-flex rounded-full h-2 w-2 bg-amber-500/70 shrink-0" title="Awaiting first report"></span>
									{/if}
									<h3 class="text-base font-bold text-slate-100 group-hover:text-white transition-colors line-clamp-1">
										{session.name}
									</h3>
								</div>
								<div class="text-[11px] text-slate-400 flex items-center gap-1.5">
									{#if session.total_reports > 0}
										<span class="text-emerald-400/90 font-medium">{formatRelativeTime(session.updated_at)}</span>
									{:else}
										<span class="text-amber-400/80 font-medium">Awaiting initial browser report</span>
									{/if}
								</div>
							</div>
							{#if confirmingDeleteId === session.id}
								<div class="flex items-center gap-1 bg-rose-950/90 border border-rose-800/90 px-2 py-1 rounded-lg animate-fade-in shrink-0">
									<span class="text-[10px] text-rose-300 font-semibold whitespace-nowrap">Delete?</span>
									<button
										onclick={() => confirmDeleteSession(session.id, session.name)}
										class="px-1.5 py-0.5 bg-rose-600 hover:bg-rose-500 text-white rounded text-[10px] font-bold transition-colors cursor-pointer"
										title="Confirm permanent deletion"
										aria-label={`Confirm permanent deletion of session "${session.name}"`}
									>
										Yes
									</button>
									<button
										onclick={cancelDeleteSession}
										class="px-1.5 py-0.5 text-slate-400 hover:text-white text-[10px] font-medium transition-colors cursor-pointer"
										title="Cancel"
										aria-label="Cancel deletion"
									>
										✕
									</button>
								</div>
							{:else}
								<button
									onclick={() => promptDeleteSession(session.id)}
									class="text-slate-500 hover:text-rose-400 p-1.5 rounded-lg hover:bg-slate-800 transition-colors shrink-0 cursor-pointer"
									title={`Delete session "${session.name}"`}
									aria-label={`Delete session "${session.name}"`}
								>
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
									</svg>
								</button>
							{/if}
						</div>

						<div class="flex items-center gap-1.5 text-xs text-slate-400 font-mono bg-slate-950 px-2.5 py-1.5 rounded-lg border border-slate-800/80 overflow-hidden text-ellipsis">
							<svg class="w-3.5 h-3.5 text-slate-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
							</svg>
							<span class="truncate">{session.target_origin}</span>
						</div>

						{#if session.description}
							<p class="text-xs text-slate-400 line-clamp-2">{session.description}</p>
						{/if}

						<div class="grid grid-cols-4 gap-2 pt-2 border-t border-slate-800/70 text-center">
							<div class="bg-slate-950/60 p-2 rounded-lg border border-slate-800/50">
								<div class="text-sm font-bold text-white">{session.total_reports || 0}</div>
								<div class="text-[10px] text-slate-400">Reports</div>
							</div>
							<div class="bg-slate-950/60 p-2 rounded-lg border border-slate-800/50">
								<div class="text-sm font-bold text-amber-400">{session.pending_violations || 0}</div>
								<div class="text-[10px] text-slate-400">Pending</div>
							</div>
							<div class="bg-slate-950/60 p-2 rounded-lg border border-slate-800/50">
								<div class="text-sm font-bold text-sky-400">{session.self_violations || 0}</div>
								<div class="text-[10px] text-slate-400">'self'</div>
							</div>
							<div class="bg-slate-950/60 p-2 rounded-lg border border-slate-800/50">
								<div class="text-sm font-bold text-emerald-400">{session.approved_rules || 0}</div>
								<div class="text-[10px] text-slate-400">Approved</div>
							</div>
						</div>
					</div>

					<div class="pt-4 mt-4 border-t border-slate-800 flex items-center justify-between gap-2">
						<button
							onclick={() => openHeaderGuide(session)}
							class="px-2.5 py-1.5 text-xs text-slate-400 hover:text-white bg-slate-800/60 hover:bg-slate-800 rounded-lg font-medium transition-colors flex items-center gap-1.5"
						>
							<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
							</svg>
							<span>Setup Header</span>
						</button>

						<a
							href={`/session/${session.id}`}
							class="px-3.5 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold shadow transition-colors flex items-center gap-1.5"
						>
							<span>Triage & Policy</span>
							<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
							</svg>
						</a>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Modal Dialogs -->
<NewSessionModal bind:open={showNewModal} onCreated={handleSessionCreated} />
<HeaderModal bind:open={showHeaderModal} session={activeSessionForHeader} />
