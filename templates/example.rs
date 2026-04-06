use std::collections::HashMap;
use std::fmt;

/// Maximum number of entries allowed in the cache.
const MAX_ENTRIES: usize = 1024;

/// Represents an error that can occur during cache operations.
#[derive(Debug)]
enum CacheError {
    NotFound(String),
    Overflow { capacity: usize, requested: usize },
}

impl fmt::Display for CacheError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            CacheError::NotFound(key) => write!(f, "key not found: {}", key),
            CacheError::Overflow { capacity, requested } => {
                write!(f, "cache overflow: {}/{}", requested, capacity)
            }
        }
    }
}

/// A generic cache with TTL support.
struct Cache<V: Clone> {
    entries: HashMap<String, V>,
    max_size: usize,
    hits: u64,
    misses: u64,
}

impl<V: Clone> Cache<V> {
    /// Creates a new cache with the given capacity.
    fn new(max_size: usize) -> Self {
        Self {
            entries: HashMap::with_capacity(max_size),
            max_size,
            hits: 0,
            misses: 0,
        }
    }

    /// Inserts a value into the cache.
    fn insert(&mut self, key: impl Into<String>, value: V) -> Result<(), CacheError> {
        if self.entries.len() >= self.max_size {
            return Err(CacheError::Overflow {
                capacity: self.max_size,
                requested: self.entries.len() + 1,
            });
        }
        self.entries.insert(key.into(), value);
        Ok(())
    }

    /// Retrieves a value from the cache.
    fn get(&mut self, key: &str) -> Option<&V> {
        if self.entries.contains_key(key) {
            self.hits += 1;
            self.entries.get(key)
        } else {
            self.misses += 1;
            None
        }
    }

    /// Returns the hit rate as a percentage.
    fn hit_rate(&self) -> f64 {
        let total = self.hits + self.misses;
        if total == 0 {
            return 0.0;
        }
        (self.hits as f64 / total as f64) * 100.0
    }
}

// Trait for items that can be serialized to a summary string.
trait Summarize {
    fn summary(&self) -> String;
}

#[derive(Clone, Debug)]
struct UserProfile {
    name: String,
    score: i32,
    active: bool,
}

impl Summarize for UserProfile {
    fn summary(&self) -> String {
        let status = if self.active { "active" } else { "inactive" };
        format!("{} (score: {}, {})", self.name, self.score, status)
    }
}

fn print_summaries<T: Summarize>(items: &[T]) {
    for item in items {
        println!("  - {}", item.summary());
    }
}

fn main() {
    let mut cache: Cache<UserProfile> = Cache::new(MAX_ENTRIES);

    let profiles = vec![
        UserProfile { name: "Alice".into(), score: 95, active: true },
        UserProfile { name: "Bob".into(), score: 82, active: false },
        UserProfile { name: "Charlie".into(), score: 73, active: true },
    ];

    for profile in &profiles {
        if let Err(e) = cache.insert(&profile.name, profile.clone()) {
            eprintln!("Error: {}", e);
        }
    }

    // Lookup and display
    match cache.get("Alice") {
        Some(user) => println!("Found: {:?}", user),
        None => println!("Not found"),
    }

    let _ = cache.get("Unknown"); // miss

    println!("Hit rate: {:.1}%", cache.hit_rate());
    println!("All profiles:");
    print_summaries(&profiles);
}
