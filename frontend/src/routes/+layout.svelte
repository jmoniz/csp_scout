<script>
	import './layout.css';
	import { onMount, onDestroy } from 'svelte';
	import Toast from '$lib/components/Toast.svelte';

	let { children } = $props();

	let backendOnline = $state(false);
	let checkingStatus = $state(true);
	let timer = null;

	async function checkHealth() {
		try {
			const controller = new AbortController();
			const timeoutId = setTimeout(() => controller.abort(), 2000);

			const res = await fetch('/health?_t=' + Date.now(), {
				signal: controller.signal,
				headers: { 'Cache-Control': 'no-cache' }
			});
			clearTimeout(timeoutId);

			if (res.ok) {
				const data = await res.json();
				backendOnline = Boolean(data && data.status === 'ok');
			} else {
				backendOnline = false;
			}
		} catch {
			backendOnline = false;
		} finally {
			checkingStatus = false;
		}
	}

	let cleanupVisibility = null;

	onMount(() => {
		checkHealth();
		// Poll status every 30 seconds (skip when tab is hidden to avoid server log spam)
		timer = setInterval(() => {
			if (typeof document !== 'undefined' && document.visibilityState === 'hidden') return;
			checkHealth();
		}, 30000);

		const handleVisibilityChange = () => {
			if (document.visibilityState === 'visible') {
				checkHealth();
			}
		};
		document.addEventListener('visibilitychange', handleVisibilityChange);
		cleanupVisibility = () => document.removeEventListener('visibilitychange', handleVisibilityChange);
	});

	onDestroy(() => {
		if (timer) clearInterval(timer);
		if (cleanupVisibility) cleanupVisibility();
	});
</script>

<svelte:head>
	<title>CSP Scout - CSP Collector & Policy Generator</title>
	<meta name="description" content="Capture Content-Security-Policy-Report-Only violations, triage blocked origins, and generate production-ready CSP rules." />
</svelte:head>

<div class="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans selection:bg-indigo-500 selection:text-white">
	<!-- Top Navigation Bar -->
	<header class="border-b border-slate-800 bg-slate-900/80 backdrop-blur sticky top-0 z-40">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
			<div class="flex items-center gap-3">
				<a href="/" class="flex items-center gap-2.5 group">
					<div class="w-9 h-9 rounded-lg bg-indigo-600 border border-indigo-500/40 flex items-center justify-center shadow-sm group-hover:bg-indigo-500 transition-colors">
						<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
						</svg>
					</div>
					<div>
						<div class="font-bold text-base tracking-tight flex items-center gap-2">
							<span>CSP Scout</span>
							<span class="text-[10px] font-mono uppercase bg-indigo-500/10 text-indigo-400 px-1.5 py-0.5 rounded border border-indigo-500/20">v1.0</span>
						</div>
						<p class="text-xs text-slate-400 -mt-0.5 hidden sm:block">CSP Collector & Policy Generator</p>
					</div>
				</a>
			</div>

			<div class="flex items-center gap-4">
				<!-- Backend status indicator -->
				<button
					type="button"
					onclick={checkHealth}
					aria-live="polite"
					class="flex items-center gap-2 text-xs px-2.5 py-1 rounded-full border border-slate-800 bg-slate-900 hover:border-slate-700 transition-all cursor-pointer shadow-sm active:scale-95"
					title="Click to re-check backend status"
				>
					{#if checkingStatus}
						<span class="w-2 h-2 rounded-full bg-amber-400 animate-pulse"></span>
						<span class="text-slate-400">Checking...</span>
					{:else if backendOnline}
						<span class="w-2 h-2 rounded-full bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.6)]"></span>
						<span class="text-emerald-400 font-medium">API Online</span>
					{:else}
						<span class="w-2 h-2 rounded-full bg-rose-400 shadow-[0_0_8px_rgba(251,113,133,0.6)]"></span>
						<span class="text-rose-400 font-medium">API Offline</span>
					{/if}
				</button>

				<a
					href="/"
					class="text-xs sm:text-sm font-medium text-slate-300 hover:text-white px-3 py-1.5 rounded-lg hover:bg-slate-800 transition-colors"
				>
					Sessions
				</a>
			</div>
		</div>
	</header>

	<!-- Main Content Outlet -->
	<main class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
		{@render children()}
	</main>

	<!-- Global Toast Overlay -->
	<Toast />

	<!-- Footer -->
	<footer class="border-t border-slate-900 bg-slate-950 py-4 text-center text-xs text-slate-500">
		<p>CSP Scout — Compliant with W3C Content Security Policy Level 2 & 3</p>
	</footer>
</div>
