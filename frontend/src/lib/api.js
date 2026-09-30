/**
 * Centralized API client for CSP Report Collector backend.
 */

const BASE_URL = '';

async function request(endpoint, options = {}) {
	const url = `${BASE_URL}${endpoint}`;
	const headers = {
		'Content-Type': 'application/json',
		...options.headers
	};

	const config = {
		...options,
		headers
	};

	const response = await fetch(url, config);

	if (response.status === 204) {
		return null;
	}

	const contentType = response.headers.get('content-type') || '';
	let data;
	if (contentType.includes('application/json')) {
		data = await response.json();
	} else {
		data = await response.text();
	}

	if (!response.ok) {
		const message = (data && data.error && data.error.message) || (typeof data === 'string' && data) || response.statusText || 'Request failed';
		const error = new Error(message);
		error.status = response.status;
		error.details = data && data.error;
		throw error;
	}

	return data;
}

export const api = {
	async getHealth() {
		return request('/health');
	},

	async getSessions() {
		return request('/api/sessions');
	},

	async createSession({ name, target_origin, description }) {
		return request('/api/sessions', {
			method: 'POST',
			body: JSON.stringify({ name, target_origin, description })
		});
	},

	async getSession(id) {
		return request(`/api/sessions/${id}`);
	},

	async deleteSession(id) {
		return request(`/api/sessions/${id}`, {
			method: 'DELETE'
		});
	},

	async getViolations(sessionId, { directive = '', status = '', search = '', source = '' } = {}) {
		const params = new URLSearchParams();
		if (directive && directive !== 'all') params.append('directive', directive);
		if (status && status !== 'all') params.append('status', status);
		if (search) params.append('search', search);
		if (source && source !== 'all') params.append('source', source);

		const qs = params.toString();
		return request(`/api/sessions/${sessionId}/violations${qs ? `?${qs}` : ''}`);
	},

	async approveAllSelf(sessionId) {
		return request(`/api/sessions/${sessionId}/violations/approve-self`, {
			method: 'POST'
		});
	},

	async getViolationSamples(sessionId, violationId, limit = 10) {
		return request(`/api/sessions/${sessionId}/violations/${violationId}/samples?limit=${limit}`);
	},

	async updateViolationStatus(sessionId, violationId, status) {
		return request(`/api/sessions/${sessionId}/violations/${violationId}`, {
			method: 'PATCH',
			body: JSON.stringify({ status })
		});
	},

	async bulkUpdateViolations(sessionId, violationIds, status) {
		return request(`/api/sessions/${sessionId}/violations/bulk`, {
			method: 'POST',
			body: JSON.stringify({ violation_ids: violationIds, status })
		});
	},

	async getPolicy(sessionId, format = '') {
		const endpoint = format ? `/api/sessions/${sessionId}/policy?format=${encodeURIComponent(format)}` : `/api/sessions/${sessionId}/policy`;
		return request(endpoint);
	},

	async updateSettings(sessionId, settings) {
		return request(`/api/sessions/${sessionId}/settings`, {
			method: 'PUT',
			body: JSON.stringify(settings)
		});
	}
};
