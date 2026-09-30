<script>
	import { toast } from '$lib/toast.js';
</script>

<div class="fixed bottom-5 right-5 z-50 flex flex-col gap-2 max-w-sm w-full pointer-events-none">
	{#each $toast as item (item.id)}
		<div
			class="pointer-events-auto flex items-center justify-between p-4 rounded-lg shadow-lg text-sm font-medium transition-all transform duration-200 border
			{item.type === 'success' ? 'bg-emerald-900/90 text-emerald-100 border-emerald-700' :
			 item.type === 'error' ? 'bg-rose-900/90 text-rose-100 border-rose-700' :
			 'bg-slate-800/90 text-slate-100 border-slate-700'}"
		>
			<div class="flex items-center gap-2">
				{#if item.type === 'success'}
					<svg class="w-5 h-5 text-emerald-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				{:else if item.type === 'error'}
					<svg class="w-5 h-5 text-rose-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				{:else}
					<svg class="w-5 h-5 text-blue-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				{/if}
				<span>{item.message}</span>
			</div>
		</div>
	{/each}
</div>
