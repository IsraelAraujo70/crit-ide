/**
 * Example TypeScript file showcasing type system and modern features.
 */

interface ApiResponse<T> {
  data: T;
  status: number;
  message: string;
  timestamp: string;
}

type HttpMethod = "GET" | "POST" | "PUT" | "DELETE";

enum LogLevel {
  Debug = "DEBUG",
  Info = "INFO",
  Warn = "WARN",
  Error = "ERROR",
}

interface User {
  readonly id: string;
  name: string;
  email: string;
  role: "admin" | "user" | "guest";
  metadata?: Record<string, unknown>;
}

// Generic repository pattern
abstract class Repository<T extends { id: string }> {
  protected items: Map<string, T> = new Map();

  abstract validate(item: T): boolean;

  async create(item: T): Promise<ApiResponse<T>> {
    if (!this.validate(item)) {
      throw new Error("Validation failed");
    }
    this.items.set(item.id, item);
    return {
      data: item,
      status: 200,
      message: "Created",
      timestamp: new Date().toISOString(),
    };
  }

  findById(id: string): T | undefined {
    return this.items.get(id);
  }

  findAll(predicate?: (item: T) => boolean): T[] {
    const all = Array.from(this.items.values());
    return predicate ? all.filter(predicate) : all;
  }
}

class UserRepository extends Repository<User> {
  validate(user: User): boolean {
    return user.name.length > 0 && user.email.includes("@");
  }

  findByRole(role: User["role"]): User[] {
    return this.findAll((u) => u.role === role);
  }
}

// Utility types
type CreateUserInput = Omit<User, "id"> & { password: string };
type UserSummary = Pick<User, "id" | "name" | "role">;

function createUser(input: CreateUserInput): User {
  const id = crypto.randomUUID();
  const { password: _, ...userData } = input;
  return { id, ...userData };
}

// Async generator
async function* paginate<T>(
  fetcher: (page: number) => Promise<T[]>,
  maxPages: number = 10
): AsyncGenerator<T[], void, unknown> {
  for (let page = 0; page < maxPages; page++) {
    const items = await fetcher(page);
    if (items.length === 0) break;
    yield items;
  }
}

// Satisfies operator and const assertions
const PERMISSIONS = {
  read: 0b001,
  write: 0b010,
  admin: 0b111,
} as const satisfies Record<string, number>;

export { UserRepository, createUser, paginate, LogLevel, PERMISSIONS };
export type { User, ApiResponse, UserSummary };
