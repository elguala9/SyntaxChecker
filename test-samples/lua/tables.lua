local Account = {}
Account.__index = Account

function Account.new(balance)
  return setmetatable({ balance = balance or 0 }, Account)
end

function Account:deposit(amount)
  self.balance = self.balance + amount
  return self.balance
end

local config = {
  name = "main",
  retries = 3,
  endpoints = { "a", "b", "c" },
  nested = { enabled = true },
}

local acc = Account.new(100)
acc:deposit(50)
for _, endpoint in ipairs(config.endpoints) do
  print(endpoint, acc.balance)
end
