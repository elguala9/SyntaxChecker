enum Role {
  Admin = "admin",
  User = "user",
}

type Result<T> =
  | { ok: true; value: T }
  | { ok: false; error: string };

class Box<T> {
  constructor(private readonly value: T) {}

  map<U>(fn: (value: T) => U): Box<U> {
    return new Box(fn(this.value));
  }

  get(): T {
    return this.value;
  }
}

function unwrap<T>(result: Result<T>): T {
  if (result.ok) {
    return result.value;
  }
  throw new Error(result.error);
}

const role: Role = Role.Admin;
const boxed = new Box(21).map((n) => n * 2);
console.log(role, unwrap({ ok: true, value: boxed.get() }));
