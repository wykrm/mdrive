# Answers

## Q1: validation order

- A returns `ErrWalletFrozen` because it checks wallet state first.
- B returns `ErrInvalidAmount` because it validates `-50` first.
- B is correct. Reject bad input before checking whether the operation is allowed. The caller has to fix the amount whether the wallet is frozen or not, and state shouldn't mask invalid input.

## Q2: meaning of `Required`

`Required` = `150`: what the caller asked for, not the shortfall.

- If `Required` is the requested amount: "Cannot withdraw 150 cents: only 100 cents is available."
- If `Required` is the 50-cent deficit: "Insufficient balance: 50 more cents is needed; 100 cents is available."

Requested amount is better. `Required=150` and `Available=100` give you both facts right away, and nobody has to guess what `Required` means. The UI can compute the 50-cent gap itself if it wants to.

## Q3: returning domain errors from the use-case layer

The use-case layer should return domain errors as-is. They're already a meaningful result of the business operation, and a prefix like `"failed"` adds nothing useful, ties callers to use-case wording, and can leak internal phrasing to the outside instead of getting properly translated.

Technically, `%w` doesn't break `errors.Is` or `errors.As`, so that's not the reason. It's the layer contract: domain outcomes pass through unchanged, while infrastructure errors can be wrapped when it adds real context. The outer boundary (HTTP, RPC, CLI, UI) then maps the domain error to the right response and message.