"use strict";

// Duplicate parameter names are a syntax error in module/strict code.
export function add(a, a) {
  return a + a;
}
