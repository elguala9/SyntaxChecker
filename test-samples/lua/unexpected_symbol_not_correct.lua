local t = { 1, 2, 3 }

-- Lua has no "++" operator; this is a syntax error.
for i = 1, #t do
  t[i]++
end
