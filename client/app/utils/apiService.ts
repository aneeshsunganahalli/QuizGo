import { authFetch, getAuthToken } from "./authUtils";

const API_BASE_URL = 'http://localhost:8080';

/**
 * API service for FlashQuiz app
 */
export const ApiService = {
  /**
   * Authentication related API calls
   */
  auth: {
    /**
     * Register a new user
     * @param {object} userData - User registration data
     * @returns {Promise} Promise with registration response
     */
    register: async (userData: { username: string, email: string, password: string }) => {
      const response = await fetch(`${API_BASE_URL}/auth/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(userData),
      });

      return response.json();
    },    /**
     * Login user
     * @param {object} credentials - User login credentials
     * @returns {Promise} Promise with login response
     */
    login: async (credentials: { email: string, password: string }) => {
      const response = await fetch(`${API_BASE_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(credentials),
      });

      // Return the response JSON with the token field
      return response.json();
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
      return authFetch(`${API_BASE_URL}/api/decks?include_public=${includePublic}`)
        .then(res => res.json());
    },

    /**
     * Get a specific deck by ID
     * @param {number} id - Deck ID
     * @returns {Promise} Promise with deck data
     */
    getById: async (id: number) => {
      return authFetch(`${API_BASE_URL}/api/decks/${id}`)
        .then(res => res.json());
    },

    /**
     * Create a new deck
     * @param {object} deckData - New deck data
     * @returns {Promise} Promise with created deck data
     */
    create: async (deckData: { title: string, description: string, category: string, isPublic: boolean }) => {
      return authFetch(`${API_BASE_URL}/api/decks`, {
        method: 'POST',
        body: JSON.stringify(deckData),
      }).then(res => res.json());
    },

    /**
     * Update an existing deck
     * @param {number} id - Deck ID
     * @param {object} deckData - Updated deck data
     * @returns {Promise} Promise with updated deck data
     */
    update: async (id: number, deckData: { title?: string, description?: string, category?: string, isPublic?: boolean }) => {
      return authFetch(`${API_BASE_URL}/api/decks/${id}`, {
        method: 'PUT',
        body: JSON.stringify(deckData),
      }).then(res => res.json());
    },

    /**
     * Delete a deck
     * @param {number} id - Deck ID
     * @returns {Promise} Promise with deletion response
     */
    delete: async (id: number) => {
      return authFetch(`${API_BASE_URL}/api/decks/${id}`, {
        method: 'DELETE',
      }).then(res => res.json());
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
      return authFetch(`${API_BASE_URL}/api/cards/deck/${deckId}`)
        .then(res => res.json());
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
      return authFetch(`${API_BASE_URL}/api/cards`, {
        method: 'POST',
        body: JSON.stringify(cardData),
      }).then(res => res.json());
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
      return authFetch(`${API_BASE_URL}/api/cards/${id}`, {
        method: 'PUT',
        body: JSON.stringify(cardData),
      }).then(res => res.json());
    },

    /**
     * Delete a card
     * @param {number} id - Card ID
     * @returns {Promise} Promise with deletion response
     */
    delete: async (id: number) => {
      return authFetch(`${API_BASE_URL}/api/cards/${id}`, {
        method: 'DELETE',
      }).then(res => res.json());
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
      return authFetch(`${API_BASE_URL}/api/study/next-cards`, {
        method: 'POST',
        body: JSON.stringify({ deck_id: deckId, limit }),
      }).then(res => res.json());
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
      return authFetch(`${API_BASE_URL}/api/study/update-progress`, {
        method: 'POST',
        body: JSON.stringify(progressData),
      }).then(res => res.json());
    },

    /**
     * Get study statistics
     * @param {number} deckId - Optional deck ID to filter stats
     * @returns {Promise} Promise with stats data
     */
    getStats: async (deckId?: number) => {
      const url = deckId 
        ? `${API_BASE_URL}/api/study/stats?deck_id=${deckId}`
        : `${API_BASE_URL}/api/study/stats`;
        
      return authFetch(url).then(res => res.json());
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
      return authFetch(`${API_BASE_URL}/api/quizzes`, {
        method: 'POST',
        body: JSON.stringify(quizData),
      }).then(res => res.json());
    },

    /**
     * Get a specific quiz
     * @param {number} id - Quiz ID
     * @returns {Promise} Promise with quiz data
     */
    getById: async (id: number) => {
      return authFetch(`${API_BASE_URL}/api/quizzes/${id}`)
        .then(res => res.json());
    },

    /**
     * Get all quizzes for the current user
     * @returns {Promise} Promise with quizzes data
     */
    getAll: async () => {
      return authFetch(`${API_BASE_URL}/api/quizzes`)
        .then(res => res.json());
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
      return authFetch(`${API_BASE_URL}/api/quizzes/answer`, {
        method: 'POST',
        body: JSON.stringify(answerData),
      }).then(res => res.json());
    },

    /**
     * Complete a quiz
     * @param {number} quizId - Quiz ID
     * @returns {Promise} Promise with quiz completion response
     */
    complete: async (quizId: number) => {
      return authFetch(`${API_BASE_URL}/api/quizzes/complete`, {
        method: 'POST',
        body: JSON.stringify({ quiz_id: quizId }),
      }).then(res => res.json());
    },
  },
};
