<script>
	import { api } from '$lib/api.js';
	import { toast } from '$lib/toast.js';

	let { open = $bindable(false), onCreated = () => {} } = $props();

	let name = $state('');
	let targetOrigin = $state('');
	let description = $state('');
	let loading = $state(false);
	let errorMessage = $state('');

	async function handleSubmit(e) {
		e.preventDefault();
		errorMessage = '';

		if (!name.trim()) {
			errorMessage = 'Session name is required.';
			return;
		}
		let origin = targetOrigin.trim();
		if (!origin) {
			errorMessage = 'Target origin URL is required.';
			return;
		}

		if (!origin.includes('://')) {
			origin = 'https://' + origin;
		}

		try {
			const parsed = new URL(origin);
			if (!['http:', 'https:'].includes(parsed.protocol)) {
				errorMessage = 'Origin must use http:// or https:// protocol.';
				return;
			}
			origin = parsed.origin;
		} catch {
			errorMessage = 'Invalid URL format for target origin (e.g. https://app.example.com).';
			return;
		}

		loading = true;
		try {
			const newSession = await api.createSession({
				name: name.trim(),
				target_origin: origin,
				description: description.trim()
			});
			toast.success(`Session "${newSession.name}" created!`);
			open = false;
			name = '';
			targetOrigin = '';
			description = '';
			onCreated(newSession);
		} catch (err) {
			errorMessage = err.message || 'Failed to create session.';
			toast.error(errorMessage);
		} finally {
			loading = false;
		}
	}

	function close() {
		if (!loading) {
			open = false;
			errorMessage = '';
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
			const nameInput = node.querySelector('#session-name');
			if (nameInput) nameInput.focus();
			else first?.focus();
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

{#if open}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-fade-in"
		role="dialog"
		aria-modal="true"
		aria-labelledby="new-session-modal-title"
		tabindex="-1"
		onclick={(e) => { if (e.target === e.currentTarget) close(); }}
		onkeydown={(e) => { if (e.key === 'Escape') close(); }}
	>
		<div
			use:trapFocus
			class="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl relative"
		>
			<button
				onclick={close}
				disabled={loading}
				class="absolute top-5 right-5 text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition-colors disabled:opacity-50"
				aria-label="Close"
			>
				<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>

			<div class="flex items-center gap-3 mb-5">
				<div class="w-10 h-10 rounded-xl bg-indigo-600/20 text-indigo-400 border border-indigo-500/30 flex items-center justify-center shrink-0">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
				</div>
				<div>
					<h3 id="new-session-modal-title" class="text-lg font-bold text-white">Create Analysis Session</h3>
					<p class="text-xs text-slate-400">Target a specific origin to collect and analyze CSP reports.</p>
				</div>
			</div>

			{#if errorMessage}
				<div class="mb-4 p-3 bg-rose-950/60 border border-rose-800 rounded-lg text-xs text-rose-300 flex items-center gap-2">
					<svg class="w-4 h-4 text-rose-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					<span>{errorMessage}</span>
				</div>
			{/if}

			<form onsubmit={handleSubmit} class="space-y-4">
				<div>
					<label for="session-name" class="block text-xs font-semibold text-slate-300 mb-1">
						Session Name <span class="text-rose-400">*</span>
					</label>
					<input
						id="session-name"
						type="text"
						bind:value={name}
						placeholder="e.g. Production Web Portal"
						required
						class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors"
					/>
				</div>

				<div>
					<label for="target-origin" class="block text-xs font-semibold text-slate-300 mb-1">
						Target Origin URL <span class="text-rose-400">*</span>
					</label>
					<input
						id="target-origin"
						type="text"
						bind:value={targetOrigin}
						placeholder="e.g. https://app.example.com"
						required
						class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors font-mono"
					/>
					<p class="text-[11px] text-slate-500 mt-1">
						Used to detect <code class="text-slate-400">'self'</code> matches and derive subdomain wildcard rules.
					</p>
				</div>

				<div>
					<label for="session-desc" class="block text-xs font-semibold text-slate-300 mb-1">
						Description (Optional)
					</label>
					<textarea
						id="session-desc"
						bind:value={description}
						rows="2"
						placeholder="e.g. Client application migration to strict CSP Report-Only mode"
						class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors"
					></textarea>
				</div>

				<div class="pt-3 flex items-center justify-end gap-3 border-t border-slate-800">
					<button
						type="button"
						onclick={close}
						disabled={loading}
						class="px-4 py-2 text-xs font-medium text-slate-400 hover:text-white rounded-lg hover:bg-slate-800 transition-colors"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={loading}
						class="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:bg-indigo-800 text-white rounded-lg text-xs font-semibold shadow-lg shadow-indigo-600/20 flex items-center gap-2 transition-all"
					>
						{#if loading}
							<svg class="animate-spin h-3.5 w-3.5 text-white" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
							<span>Creating...</span>
						{:else}
							<span>Create Session</span>
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
