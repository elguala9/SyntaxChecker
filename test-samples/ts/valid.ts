interface User {
  id: number;
  name: string;
}

function greet(user: User): string {
  return `hello, ${user.name}`;
}

const u: User = { id: 1, name: "world" };
console.log(greet(u));
