export function Broken() {
  // The <ul> element is never closed.
  return (
    <ul>
      <li>one</li>
      <li>two</li>
  );
}
