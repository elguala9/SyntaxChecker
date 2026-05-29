export class Queue {
  #items = [];

  enqueue(item) {
    this.#items.push(item);
    return this;
  }

  async drain(handler) {
    while (this.#items.length > 0) {
      const item = this.#items.shift();
      await handler(item);
    }
  }

  get size() {
    return this.#items.length;
  }
}

const q = new Queue();
q.enqueue(1).enqueue(2);
await q.drain(async (n) => console.log(n));
