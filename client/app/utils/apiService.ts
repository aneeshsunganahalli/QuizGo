import { getAuthToken } from "./authUtils";
import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080';

// Create axios instance with base configuration
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  }
});

// Create authenticated axios instance
const authApi = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  }
});

// Add interceptor to add auth token to requests
authApi.interceptors.request.use((config) => {
  const token = getAuthToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

/**
 * API service for FlashQuiz app
 */
export const ApiService = {  /**
   * Authentication related API calls
   */
  auth: {
    /**
     * Register a new user
     * @param {object} userData - User registration data
     * @returns {Promise} Promise with registration response
     */
    register: async (userData: { username: string, email: string, password: string }) => {
      const response = await api.post('/auth/register', userData);
      return response.data;
    },    /**
     * Login user
     * @param {object} credentials - User login credentials
     * @returns {Promise} Promise with login response
     */
    login: async (credentials: { username: string, password: string }) => {
      const response = await api.post('/auth/login', credentials);
      
      // Return the response data with the token field
      return response.data;
    },
  },
  /**
   * Decks related API calls
   */
  decks: {
    /**
     * Get all decks for the current user
     * @param {boolean} includePublic - Whether to include public decks
     * @returns {Promise} Promise with decks data
     */
    getAll: async (includePublic = false) => {
      const response = await authApi.get(`/api/decks?include_public=${includePublic}`);
      return response.data;
    },

    /**
     * Get a specific deck by ID
     * @param {number} id - Deck ID
     * @returns {Promise} Promise with deck data
     */
    getById: async (id: number) => {
      const response = await authApi.get(`/api/decks/${id}`);
      return response.data;
    },

    /**
     * Create a new deck
     * @param {object} deckData - New deck data
     * @returns {Promise} Promise with created deck data
     */
    create: async (deckData: { title: string, description: string, category: string, isPublic: boolean }) => {
      const response = await authApi.post('/api/decks', deckData);
      return response.data;
    },

    /**
     * Update an existing deck
     * @param {number} id - Deck ID
     * @param {object} deckData - Updated deck data
     * @returns {Promise} Promise with updated deck data
     */
    update: async (id: number, deckData: { title?: string, description?: string, category?: string, isPublic?: boolean }) => {
      const response = await authApi.put(`/api/decks/${id}`, deckData);
      return response.data;
    },

    /**
     * Delete a deck
     * @param {number} id - Deck ID
     * @returns {Promise} Promise with deletion response
     */
    delete: async (id: number) => {
      const response = await authApi.delete(`/api/decks/${id}`);
      return response.data;
    },
  },
  /**
   * Cards related API calls
   */
  cards: {
    /**
     * Get all cards in a deck
     * @param {number} deckId - Deck ID
     * @returns {Promise} Promise with cards data
     */
    getByDeck: async (deckId: number) => {
      const response = await authApi.get(`/api/cards/deck/${deckId}`);
      return response.data;
    },

    /**
     * Create a new card
     * @param {object} cardData - New card data
     * @returns {Promise} Promise with created card data
     */
    create: async (cardData: { 
      deckId: number, 
      frontContent: string, 
      backContent: string, 
      contentType?: string, 
      difficultyLevel?: number 
    }) => {
      const response = await authApi.post('/api/cards', cardData);
      return response.data;
    },

    /**
     * Update an existing card
     * @param {number} id - Card ID
     * @param {object} cardData - Updated card data
     * @returns {Promise} Promise with updated card data
     */
    update: async (id: number, cardData: { 
      frontContent?: string, 
      backContent?: string, 
      contentType?: string, 
      difficultyLevel?: number 
    }) => {
      const response = await authApi.put(`/api/cards/${id}`, cardData);
      return response.data;
    },

    /**
     * Delete a card
     * @param {number} id - Card ID
     * @returns {Promise} Promise with deletion response
     */
    delete: async (id: number) => {
      const response = await authApi.delete(`/api/cards/${id}`);
      return response.data;
    },
  },
  /**
   * Study related API calls
   */
  study: {
    /**
     * Get next cards for study
     * @param {number} deckId - Deck ID
     * @param {number} limit - Maximum number of cards to return
     * @returns {Promise} Promise with cards data
     */
    getNextCards: async (deckId: number, limit = 20) => {
      const response = await authApi.post('/api/study/next-cards', { deck_id: deckId, limit });
      return response.data;
    },

    /**
     * Update card progress
     * @param {object} progressData - Card progress data
     * @returns {Promise} Promise with updated progress data
     */
    updateProgress: async (progressData: { 
      cardId: number, 
      performance: number, // 1-5 scale
      timeSpent?: number 
    }) => {
      const response = await authApi.post('/api/study/update-progress', progressData);
      return response.data;
    },

    /**
     * Get study statistics
     * @param {number} deckId - Optional deck ID to filter stats
     * @returns {Promise} Promise with stats data
     */
    getStats: async (deckId?: number) => {
      const url = deckId 
        ? `/api/study/stats?deck_id=${deckId}`
        : `/api/study/stats`;
        
      const response = await authApi.get(url);
      return response.data;
    },
  },
  /**
   * Quiz related API calls
   */
  quizzes: {
    /**
     * Create a new quiz
     * @param {object} quizData - New quiz data
     * @returns {Promise} Promise with created quiz data
     */
    create: async (quizData: { 
      deckId: number, 
      title: string, 
      description: string,
      cardCount?: number
    }) => {
      const response = await authApi.post('/api/quizzes', quizData);
      return response.data;
    },

    /**
     * Get a specific quiz
     * @param {number} id - Quiz ID
     * @returns {Promise} Promise with quiz data
     */
    getById: async (id: number) => {
      const response = await authApi.get(`/api/quizzes/${id}`);
      return response.data;
    },

    /**
     * Get all quizzes for the current user
     * @returns {Promise} Promise with quizzes data
     */
    getAll: async () => {
      const response = await authApi.get('/api/quizzes');
      return response.data;
    },

    /**
     * Submit an answer to a quiz question
     * @param {object} answerData - Answer data
     * @returns {Promise} Promise with answer submission response
     */
    submitAnswer: async (answerData: { 
      questionId: number, 
      answer: string,
      timeSpent?: number
    }) => {
      const response = await authApi.post('/api/quizzes/answer', answerData);
      return response.data;
    },

    /**
     * Complete a quiz
     * @param {number} quizId - Quiz ID
     * @returns {Promise} Promise with quiz completion response
     */
    complete: async (quizId: number) => {
      const response = await authApi.post('/api/quizzes/complete', { quiz_id: quizId });
      return response.data;
    },
  },
};
