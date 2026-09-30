import { writable } from 'svelte/store';

function createToastStore() {
	const { subscribe, update } = writable([]);

	return {
		subscribe,
		show(message, type = 'info', duration = 3000) {
			const id = Math.random().toString(36).substring(2, 9);
			update((toasts) => [...toasts, { id, message, type }]);

			setTimeout(() => {
				update((toasts) => toasts.filter((t) => t.id !== id));
			}, duration);
		},
		success(message, duration) {
			this.show(message, 'success', duration);
		},
		error(message, duration) {
			this.show(message, 'error', duration);
		},
		info(message, duration) {
			this.show(message, 'info', duration);
		}
	};
}

export const toast = createToastStore();
