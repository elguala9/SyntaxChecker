def add_tax(rate): . * (1 + rate);

.products
  | map(.price | add_tax(0.22))
  | reduce .[] as $p (0; . + $p)
