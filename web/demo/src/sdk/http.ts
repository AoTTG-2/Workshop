export let apiBase = 'http://localhost:8080/api';

export function setApiBase(newUrl: string) {
    apiBase = newUrl;
}

export let debugUserId = '';
export let debugUserRoles: string[] = [];

export function setDebugAuth(userId: string, userRoles: string[]) {
    debugUserId = userId;
    debugUserRoles = userRoles;
}

interface RequestOptions extends RequestInit {
    params?: Record<string, any>;
}

function buildQuery(params: Record<string, any>): string {
    return Object.keys(params)
        .filter(key => params[key] !== undefined && params[key] !== null)
        .map(key => {
            return Array.isArray(params[key])
                ? params[key]
                    .map((v: string) => `${encodeURIComponent(key)}=${encodeURIComponent(v)}`)
                    .join('&')
                : `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`;
        })
        .join('&');
}

export async function callApi(endpoint: string, options: RequestOptions = {}): Promise<any> {
    let url = `${apiBase}${endpoint}`;
    if (options.params) {
        url += '?' + buildQuery(options.params);
    }
    const headers = new Headers(options.headers || {});
    if (debugUserId) {
        headers.set('X-Debug-User-ID', debugUserId);
    }
    if (debugUserRoles && debugUserRoles.length > 0) {
        headers.set('X-Debug-User-Roles', debugUserRoles.join(','));
    }
    const response = await fetch(url, {...options, headers});
    if (!response.ok) {
        throw new Error(`API error: ${response.status}`);
    }
    const text = await response.text();
    if (text.trim().length > 0) {
        try {
            return JSON.parse(text);
        } catch (e) {
            throw new Error('Failed to parse JSON response');
        }
    }
    return null;
}
