const numbers = [1, 2, 3, 4];

const doubled = numbers.map((n) => n * 2);

function sum(values) {
  return values.reduce((acc, v) => acc + v, 0);
}

console.log(sum(doubled));
