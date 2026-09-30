<script>
	import { fade } from 'svelte/transition';
	import { api } from '$lib/api.js';
	import { toast } from '$lib/toast.js';

	let { sessionId = '', targetOrigin = '', onViolationUpdated = () => {} } = $props();

	let violations = $state([]);
	let loading = $state(true);
	let error = $state('');

	let activeDirective = $state('all');
	let activeStatus = $state('pending'); // Default to Inbox / Pending Triage
	let activeSource = $state('all'); // 'all' | 'self' | 'third_party'
	let searchQuery = $state('');

	let selectedIds = $state(new Set());
	let expandedViolationId = $state(null);
	let samplesLoading = $state(false);
	let samplesData = $state([]);

	const directivesList = [
		'all',
		'connect-src',
		'script-src',
		'style-src',
		'img-src',
		'font-src',
		'frame-src',
		'worker-src',
		'media-src',
		'object-src',
		'manifest-src',
		'form-action',
		'frame-ancestors'
	];

	export async function reload(quiet = false) {
		if (!quiet) loading = true;
		error = '';
		try {
			violations = await api.getViolations(sessionId, {
				directive: activeDirective,
				status: activeStatus,
				search: searchQuery,
				source: activeSource
			});
			if (!quiet) selectedIds.clear();
		} catch (err) {
			if (!quiet) error = err.message || 'Failed to load violations';
		} finally {
			if (!quiet) loading = false;
		}
	}

	$effect(() => {
		const _id = sessionId;
		const _dir = activeDirective;
		const _stat = activeStatus;
		const _src = activeSource;
		const _q = searchQuery;
		if (_id) {
			reload();
		}
	});

	let pendingSelfCount = $derived(violations.filter((v) => v.is_self && v.status === 'pending').length);

	async function handleApproveAllSelf() {
		try {
			const res = await api.approveAllSelf(sessionId);
			toast.success(`Approved ${res.approved} 'self' violations!`);
			if (activeStatus === 'pending') {
				// Remove all 'self' items from the pending inbox queue
				violations = violations.filter((v) => !v.is_self);
				selectedIds.clear();
				selectedIds = new Set();
			} else {
				await reload();
			}
			onViolationUpdated();
		} catch (err) {
			toast.error('Failed to approve self violations: ' + err.message);
		}
	}

	async function handleStatusChange(violationId, newStatus) {
		try {
			const updated = await api.updateViolationStatus(sessionId, violationId, newStatus);
			if (activeStatus === 'pending' || (activeStatus !== 'all' && activeStatus !== newStatus)) {
				// Completed task: remove item immediately from current queue view
				violations = violations.filter((v) => v.id !== violationId);
				if (selectedIds.has(violationId)) {
					selectedIds.delete(violationId);
					selectedIds = new Set(selectedIds);
				}
				if (expandedViolationId === violationId) {
					expandedViolationId = null;
				}
			} else {
				violations = violations.map((v) => (v.id === violationId ? updated : v));
			}
			toast.success(`Rule marked as ${newStatus.replace('_', ' ')}`);
			onViolationUpdated();
		} catch (err) {
			toast.error(err.message || 'Failed to update rule');
		}
	}

	async function handleBulkAction(status) {
		if (selectedIds.size === 0) return;
		const ids = Array.from(selectedIds);
		try {
			await api.bulkUpdateViolations(sessionId, ids, status);
			toast.success(`Bulk updated ${ids.length} violations to ${status.replace('_', ' ')}`);
			selectedIds.clear();
			selectedIds = new Set();
			if (activeStatus === 'pending' || (activeStatus !== 'all' && activeStatus !== status)) {
				const idSet = new Set(ids);
				violations = violations.filter((v) => !idSet.has(v.id));
			} else {
				await reload();
			}
			onViolationUpdated();
		} catch (err) {
			toast.error(err.message || 'Bulk update failed');
		}
	}

	function toggleSelectAll() {
		if (selectedIds.size === violations.length) {
			selectedIds.clear();
		} else {
			selectedIds = new Set(violations.map((v) => v.id));
		}
	}

	function toggleSelect(id) {
		if (selectedIds.has(id)) {
			selectedIds.delete(id);
		} else {
			selectedIds.add(id);
		}
		selectedIds = new Set(selectedIds);
	}

	async function toggleExpandEvidence(violation) {
		if (expandedViolationId === violation.id) {
			expandedViolationId = null;
			samplesData = [];
			return;
		}

		expandedViolationId = violation.id;
		samplesLoading = true;
		samplesData = [];
		try {
			samplesData = await api.getViolationSamples(sessionId, violation.id, 10);
		} catch (err) {
			toast.error('Failed to load violation evidence');
		} finally {
			samplesLoading = false;
		}
	}

	function getDirectiveColor(dir) {
		switch (dir) {
			case 'connect-src': return 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20';
			case 'script-src': return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
			case 'style-src': return 'bg-fuchsia-500/10 text-fuchsia-400 border-fuchsia-500/20';
			case 'img-src': return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
			case 'font-src': return 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20';
			case 'frame-src': return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
			default: return 'bg-slate-800 text-slate-300 border-slate-700';
		}
	}

	function getStatusBadge(status) {
		switch (status) {
			case 'approved_origin':
				return { label: 'Origin Approved', color: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30' };
			case 'approved_wildcard':
				return { label: 'Wildcard Approved', color: 'bg-teal-500/20 text-teal-300 border-teal-500/30' };
			case 'approved_self':
				return { label: "'self' Approved", color: 'bg-sky-500/20 text-sky-300 border-sky-500/30' };
			case 'rejected':
				return { label: 'Rejected', color: 'bg-rose-500/20 text-rose-300 border-rose-500/30' };
			case 'ignored':
				return { label: 'Ignored', color: 'bg-slate-700/50 text-slate-400 border-slate-600' };
			default:
				return { label: 'Pending Triage', color: 'bg-amber-500/20 text-amber-300 border-amber-500/30' };
		}
	}
</script>

<div class="space-y-4">
	<!-- Directive Filter Pills -->
	<div class="flex items-center gap-1.5 overflow-x-auto pb-2 scrollbar-thin">
		{#each directivesList as dir}
			<button
				onclick={() => activeDirective = dir}
				class="px-3 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-all flex items-center gap-1.5 {activeDirective === dir ? 'bg-indigo-600 text-white shadow-sm' : 'bg-slate-900 text-slate-400 hover:text-white hover:bg-slate-800 border border-slate-800'}"
			>
				<span>{dir === 'all' ? 'All Directives' : dir}</span>
			</button>
		{/each}
	</div>

	<!-- Controls Bar: Source Filter, Status Filter, Search, Bulk Actions -->
	<div class="flex flex-col lg:flex-row items-stretch lg:items-center justify-between gap-3 bg-slate-900/60 p-3 rounded-xl border border-slate-800">
		<div class="flex flex-wrap items-center gap-2.5">
			<!-- Source Type Filter (All / 'self' / Third-party) -->
			<div class="flex items-center bg-slate-950 p-1 rounded-lg border border-slate-800 text-xs">
				<button
					type="button"
					onclick={() => activeSource = 'all'}
					class="px-2.5 py-1 rounded transition-colors {activeSource === 'all' ? 'bg-indigo-600 text-white font-medium shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
				>
					All Origins
				</button>
				<button
					type="button"
					onclick={() => activeSource = 'self'}
					class="px-2.5 py-1 rounded flex items-center gap-1.5 transition-colors {activeSource === 'self' ? 'bg-sky-600 text-white font-medium shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
				>
					<span class="w-1.5 h-1.5 rounded-full bg-sky-400"></span>
					<span>First-Party ('self')</span>
				</button>
				<button
					type="button"
					onclick={() => activeSource = 'third_party'}
					class="px-2.5 py-1 rounded transition-colors {activeSource === 'third_party' ? 'bg-indigo-600 text-white font-medium shadow-sm' : 'text-slate-400 hover:text-slate-200'}"
				>
					Third-Party
				</button>
			</div>

			<div class="flex items-center gap-1.5">
				<select
					bind:value={activeStatus}
					class="bg-slate-950 border border-slate-800 text-xs text-slate-300 rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-indigo-500 font-medium"
				>
					<option value="pending">📥 Inbox (Pending Triage)</option>
					<option value="all">All Rules</option>
					<option value="approved_origin">Approved (Origin)</option>
					<option value="approved_wildcard">Approved (Wildcard)</option>
					<option value="approved_self">Approved ('self')</option>
					<option value="rejected">Rejected</option>
					<option value="ignored">Ignored</option>
				</select>
				{#if activeStatus === 'pending' && !loading}
					<span class="px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/20 text-[11px] font-semibold whitespace-nowrap">
						{violations.length} remaining
					</span>
				{/if}
			</div>

			<div class="relative flex-1 min-w-[200px]">
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search origin or wildcard..."
					class="w-full bg-slate-950 border border-slate-800 text-xs text-slate-100 placeholder-slate-500 rounded-lg pl-8 pr-3 py-1.5 focus:outline-none focus:border-indigo-500"
				/>
				<svg class="w-3.5 h-3.5 text-slate-500 absolute left-2.5 top-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
			</div>
		</div>

		<div class="flex items-center gap-2">
			<!-- Quick 1-Click Approve All 'self' Button -->
			{#if pendingSelfCount > 0}
				<button
					type="button"
					onclick={handleApproveAllSelf}
					class="px-3 py-1.5 bg-sky-600 hover:bg-sky-500 text-white rounded-lg text-xs font-semibold shadow-md shadow-sky-600/30 flex items-center gap-1.5 transition-all transform hover:-translate-y-0.5 active:translate-y-0"
					title="Approve 'self' for all pending first-party resources at once"
				>
					<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
					</svg>
					<span>Approve All 'self' ({pendingSelfCount})</span>
				</button>
			{/if}

			{#if selectedIds.size > 0}
				<div class="flex items-center gap-2 bg-indigo-950/60 px-3 py-1.5 rounded-lg border border-indigo-800/80 animate-fade-in text-xs">
					<span class="text-indigo-300 font-semibold">{selectedIds.size} selected</span>
					<div class="h-4 w-px bg-indigo-800 mx-1"></div>
					<button
						type="button"
						onclick={() => handleBulkAction('approved_self')}
						class="px-2 py-1 bg-sky-700 hover:bg-sky-600 text-white rounded font-medium text-[11px]"
					>
						Approve 'self'
					</button>
					<button
						type="button"
						onclick={() => handleBulkAction('approved_origin')}
						class="px-2 py-1 bg-emerald-700 hover:bg-emerald-600 text-white rounded font-medium text-[11px]"
					>
						Approve Origin
					</button>
					<button
						type="button"
						onclick={() => handleBulkAction('approved_wildcard')}
						class="px-2 py-1 bg-teal-700 hover:bg-teal-600 text-white rounded font-medium text-[11px]"
					>
						Wildcard
					</button>
					<button
						type="button"
						onclick={() => handleBulkAction('rejected')}
						class="px-2 py-1 bg-rose-700 hover:bg-rose-600 text-white rounded font-medium text-[11px]"
					>
						Reject
					</button>
				</div>
			{/if}
		</div>
	</div>

	<!-- Violations Table -->
	{#if loading}
		<div class="py-16 text-center text-slate-500 flex flex-col items-center justify-center gap-3">
			<svg class="animate-spin h-6 w-6 text-indigo-500" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
			</svg>
			<p class="text-xs">Loading violation records...</p>
		</div>
	{:else if error}
		<div class="p-6 bg-rose-950/40 border border-rose-800/60 rounded-xl text-center">
			<p class="text-rose-400 text-xs font-semibold mb-2">{error}</p>
			<button
				onclick={reload}
				class="px-3 py-1 bg-rose-800 hover:bg-rose-700 text-white rounded text-xs font-medium"
			>
				Retry
			</button>
		</div>
	{:else if violations.length === 0}
		<div class="bg-slate-900/40 border border-dashed border-slate-800 rounded-xl p-8 sm:p-12 text-center space-y-3">
			{#if activeStatus === 'pending'}
				<div class="w-12 h-12 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center justify-center mx-auto">
					<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				</div>
				<h3 class="text-sm font-semibold text-white">All caught up! No pending violations</h3>
				<p class="text-xs text-slate-400 max-w-md mx-auto">
					All captured resources matching this filter have been reviewed and processed.
				</p>
				<button
					type="button"
					onclick={() => activeStatus = 'all'}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-indigo-400 hover:text-indigo-300 text-xs font-medium transition-colors border border-slate-700"
				>
					<span>View All Processed Rules</span>
					<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
				</button>
			{:else}
				<p class="text-xs text-slate-400">No violations match the current filter criteria.</p>
			{/if}
		</div>
	{:else}
		<div class="border border-slate-800 rounded-xl overflow-hidden bg-slate-900/60">
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs border-collapse">
					<thead>
						<tr class="bg-slate-950/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider text-[11px]">
							<th class="p-3 w-8">
								<input
									type="checkbox"
									checked={selectedIds.size === violations.length && violations.length > 0}
									onchange={toggleSelectAll}
									class="rounded border-slate-700 bg-slate-900 text-indigo-600 focus:ring-indigo-500 cursor-pointer"
								/>
							</th>
							<th class="p-3">Directive</th>
							<th class="p-3">Blocked Origin / Resource</th>
							<th class="p-3">Wildcard Suggestion</th>
							<th class="p-3 text-center">Reports</th>
							<th class="p-3 text-center">Status</th>
							<th class="p-3 text-right">Triage Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-slate-800/60 font-normal">
						{#each violations as v (v.id)}
							{@const badge = getStatusBadge(v.status)}
							<tr transition:fade={{ duration: 150 }} class="hover:bg-slate-800/30 transition-colors {selectedIds.has(v.id) ? 'bg-indigo-950/20' : ''}">
								<td class="p-3">
									<input
										type="checkbox"
										checked={selectedIds.has(v.id)}
										onchange={() => toggleSelect(v.id)}
										class="rounded border-slate-700 bg-slate-900 text-indigo-600 focus:ring-indigo-500 cursor-pointer"
									/>
								</td>
								<td class="p-3">
									<span class="px-2 py-0.5 rounded border text-[11px] font-mono {getDirectiveColor(v.directive)}">
										{v.directive}
									</span>
								</td>
								<td class="p-3">
									<div class="flex items-center gap-2">
										<span class="font-mono text-slate-200 select-all font-medium">{v.origin_host}</span>
										{#if v.is_self}
											<span class="px-1.5 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/20 text-[10px] font-medium" title="Origin matches session target">
												'self' match
											</span>
										{/if}
									</div>
								</td>
								<td class="p-3">
									{#if v.suggested_wildcard}
										<code class="font-mono text-slate-400 text-[11px] select-all bg-slate-950/60 px-1.5 py-0.5 rounded border border-slate-800">
											{v.suggested_wildcard}
										</code>
									{:else}
										<span class="text-slate-600 text-[11px] italic">—</span>
									{/if}
								</td>
								<td class="p-3 text-center">
									<button
										onclick={() => toggleExpandEvidence(v)}
										class="inline-flex items-center gap-1 font-bold text-slate-200 hover:text-indigo-400 transition-colors"
										title="Click to view report occurrences"
									>
										<span>{v.count}</span>
										<svg class="w-3.5 h-3.5 text-slate-500 transition-transform {expandedViolationId === v.id ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
										</svg>
									</button>
								</td>
								<td class="p-3 text-center">
									<span class="px-2 py-0.5 rounded-full border text-[10px] font-medium {badge.color}">
										{badge.label}
									</span>
								</td>
								<td class="p-3 text-right">
									<div class="flex items-center justify-end gap-1.5">
										<!-- Approve Origin Button -->
										<button
											onclick={() => handleStatusChange(v.id, 'approved_origin')}
											class="px-2 py-1 rounded bg-slate-800 hover:bg-emerald-900/60 hover:text-emerald-300 hover:border-emerald-700 text-slate-300 border border-slate-700 transition-colors text-[11px]"
											title={`Approve exact origin ${v.origin_host}`}
										>
											Approve Origin
										</button>

										<!-- Approve Wildcard Button -->
										{#if v.suggested_wildcard}
											<button
												onclick={() => handleStatusChange(v.id, 'approved_wildcard')}
												class="px-2 py-1 rounded bg-slate-800 hover:bg-teal-900/60 hover:text-teal-300 hover:border-teal-700 text-slate-300 border border-slate-700 transition-colors text-[11px]"
												title={`Approve wildcard ${v.suggested_wildcard}`}
											>
												Wildcard
											</button>
										{/if}

										<!-- Mark as 'self' Button -->
										{#if v.is_self}
											<button
												onclick={() => handleStatusChange(v.id, 'approved_self')}
												class="px-2 py-1 rounded bg-slate-800 hover:bg-sky-900/60 hover:text-sky-300 hover:border-sky-700 text-slate-300 border border-slate-700 transition-colors text-[11px]"
												title="Cover with 'self' keyword"
											>
												'self'
											</button>
										{/if}

										<!-- Reject Button -->
										<button
											onclick={() => handleStatusChange(v.id, 'rejected')}
											class="p-1 rounded text-slate-500 hover:text-rose-400 hover:bg-slate-800 transition-colors"
											title="Reject resource"
										>
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
											</svg>
										</button>
									</div>
								</td>
							</tr>

							<!-- Expanded Evidence Drawer Row -->
							{#if expandedViolationId === v.id}
								<tr class="bg-slate-950/80 border-y border-slate-800">
									<td colspan="7" class="p-4">
										<div class="space-y-3">
											<div class="flex items-center justify-between">
												<h4 class="text-xs font-bold text-slate-300 flex items-center gap-1.5">
													<span>Recent Violation Samples</span>
													<span class="text-[10px] text-slate-500">({samplesData.length} records shown)</span>
												</h4>
												<button
													onclick={() => expandedViolationId = null}
													class="text-xs text-slate-500 hover:text-slate-300"
												>
													Close
												</button>
											</div>

											{#if samplesLoading}
												<div class="py-4 text-center text-slate-500 text-xs">Loading samples...</div>
											{:else if samplesData.length === 0}
												<div class="py-3 text-center text-slate-500 text-xs">No raw samples available.</div>
											{:else}
												<div class="overflow-x-auto">
													<table class="w-full text-left text-[11px] border border-slate-800 rounded bg-slate-900/40">
														<thead>
															<tr class="bg-slate-950 text-slate-400 border-b border-slate-800">
																<th class="p-2">Document URI</th>
																<th class="p-2">Blocked URI</th>
																<th class="p-2">Source Script</th>
																<th class="p-2">Position</th>
																<th class="p-2">Time</th>
															</tr>
														</thead>
														<tbody class="divide-y divide-slate-800/60 font-mono text-slate-300">
															{#each samplesData as sample (sample.id)}
																<tr>
																	<td class="p-2 max-w-[200px] truncate" title={sample.document_uri}>
																		{sample.document_uri || '—'}
																	</td>
																	<td class="p-2 max-w-[240px] truncate text-amber-300" title={sample.blocked_uri}>
																		{sample.blocked_uri}
																	</td>
																	<td class="p-2 max-w-[180px] truncate" title={sample.source_file}>
																		{sample.source_file || '—'}
																	</td>
																	<td class="p-2">
																		{sample.line_number ? `${sample.line_number}:${sample.column_number}` : '—'}
																	</td>
																	<td class="p-2 text-slate-500 text-[10px] font-sans">
																		{new Date(sample.created_at).toLocaleTimeString()}
																	</td>
																</tr>
															{/each}
														</tbody>
													</table>
												</div>
											{/if}
										</div>
									</td>
								</tr>
							{/if}
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
