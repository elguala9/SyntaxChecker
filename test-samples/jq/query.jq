.items[]
  | select(.active)
  | {name: .name, total: (.price * .qty)}
