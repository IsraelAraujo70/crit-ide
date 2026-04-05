/**
 * Example JavaScript file for syntax highlighting demo.
 * Covers: keywords, strings, template literals, arrow functions, classes.
 */

const API_URL = "https://api.example.com/v1";
const MAX_ITEMS = 100;
let requestCount = 0;

// A simple fetch wrapper with retry logic
async function fetchWithRetry(url, options = {}, retries = 3) {
  for (let attempt = 0; attempt < retries; attempt++) {
    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          "Content-Type": "application/json",
          ...options.headers,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      requestCount++;
      return await response.json();
    } catch (error) {
      if (attempt === retries - 1) throw error;
      console.warn(`Attempt ${attempt + 1} failed, retrying...`);
      await new Promise((resolve) => setTimeout(resolve, 1000 * (attempt + 1)));
    }
  }
}

class EventEmitter {
  #listeners = new Map();

  on(event, callback) {
    if (!this.#listeners.has(event)) {
      this.#listeners.set(event, []);
    }
    this.#listeners.get(event).push(callback);
    return this;
  }

  emit(event, ...args) {
    const callbacks = this.#listeners.get(event) ?? [];
    callbacks.forEach((cb) => cb(...args));
  }
}

// Arrow function with destructuring
const formatUser = ({ name, email, age = null }) => {
  const display = age !== null ? `${name} (${age})` : name;
  return { display, email, active: true };
};

// Array methods and template literals
const users = [
  { name: "Alice", email: "alice@test.com", age: 30 },
  { name: "Bob", email: "bob@test.com", age: 25 },
  { name: "Charlie", email: "charlie@test.com" },
];

const formatted = users
  .filter((u) => u.age !== undefined)
  .map(formatUser)
  .sort((a, b) => a.display.localeCompare(b.display));

console.log(`Processed ${formatted.length} users out of ${users.length}`);

// Optional chaining and nullish coalescing
const config = {
  server: { host: "localhost", port: 3000 },
};
const port = config?.server?.port ?? 8080;
const dbHost = config?.database?.host ?? "localhost";

export { fetchWithRetry, EventEmitter, formatUser };
