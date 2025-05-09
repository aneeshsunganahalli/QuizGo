/**
 * Authentication utility functions for FlashQuiz app
 */

/**
 * Check if user is authenticated
 * @returns {boolean} True if user is authenticated, false otherwise
 */
export const isAuthenticated = (): boolean => {
  if (typeof window === 'undefined') {
    return false; // Always return false during SSR
  }
  
  const token = localStorage.getItem('authToken');
  return !!token;
};

/**
 * Get the current user data
 * @returns {object|null} User object if authenticated, null otherwise
 */
export const getCurrentUser = () => {
  if (typeof window === 'undefined') {
    return null; // Return null during SSR
  }
  
  const userJson = localStorage.getItem('user');
  if (!userJson) return null;
  
  try {
    return JSON.parse(userJson);
  } catch (err) {
    console.error('Error parsing user data:', err);
    return null;
  }
};

/**
 * Get the authentication token
 * @returns {string|null} Auth token if available, null otherwise
 */
export const getAuthToken = (): string | null => {
  if (typeof window === 'undefined') {
    return null;
  }
  
  return localStorage.getItem('authToken');
};

/**
 * Logout the current user
 */
export const logout = () => {
  if (typeof window === 'undefined') {
    return;
  }
  
  // Clear localStorage items
  localStorage.removeItem('authToken');
  localStorage.removeItem('user');
  
  // Clear the authentication cookie used by middleware
  document.cookie = 'authToken=; path=/; expires=Thu, 01 Jan 1970 00:00:01 GMT';
  
  // Reload the page to reset all states
  window.location.href = '/login';
};

/**
 * Authenticated fetch utility that includes the auth token
 * @param {string} url - The URL to fetch
 * @param {object} options - Fetch options
 * @returns {Promise} Fetch promise
 */
export const authFetch = async (url: string, options: RequestInit = {}): Promise<Response> => {
  const token = getAuthToken();
  
  const headers = {
    ...(options.headers || {}),
    'Authorization': token ? `Bearer ${token}` : '',
    'Content-Type': 'application/json',
  };
  
  return fetch(url, {
    ...options,
    headers,
  });
};
