# Recovery

Recovery is a way to add a credential to an account you do not currently control. That is its purpose and also its danger: **an account is only as secure as its easiest recovery path**, no matter how good the passkeys are.

In a passkey-only system, two situations account for nearly all of it:

- *"I have a new phone."* Usually not recovery at all — synced passkeys follow the person through their provider's own account.
- *"I lost everything."* This is the one that needs an answer decided in advance, because the alternative is inventing one under pressure for a person who is upset.

## The ladder

Portico's position, in the order a deployment should rely on them.

### 1. A second passkey (the real answer)

Prompt for a second credential right after the first enrollment, and keep prompting while an account has only one. A passkey on a second device, a second provider, or a hardware key removes the single point of failure entirely and costs nothing to operate.

Everything below this line exists because people will not do it.

### 2. Recovery codes

Ten single-use codes, generated at enrollment, shown exactly once, stored hashed like any other secret. Each one can be spent to run the recovery ceremony.

They are the best value in this document: offline, unphishable, no new channel to attack, and no dependency on a mail provider. Regenerating the set invalidates every outstanding code, and spending one leaves the others alone. Show the remaining count in account settings, because a person with two left should know.

### 3. A verified identifier

Any **verified** email or phone on the account can start the recovery ceremony. Several identifiers are better than one: the second email is what saves most people in practice.

A phone is a factor, not a door. Recovery that can be completed by SMS alone is recovery by SIM swap, so a phone should either be paired with something else or limited to lower-risk steps.

### 4. The recovery ceremony itself

Whatever gates it — a code, a link, a second identifier — the ceremony that follows is the same, and it is deliberately narrow:

- It is **a registration ceremony that attaches a credential to an existing subject**.
- It creates no account, revokes nothing, and changes no other state.
- It uses a longer window than enrollment (15 minutes against 5), because it starts from a message read on another device.
- It is **never gated by whether registration is open.** A closed deployment that gates recovery locks its own users out permanently, and leaves the first administrator unable to enroll at all.
- It does **not** automatically revoke existing credentials. The legitimate holder may still have a working passkey, and automatic revocation hands an attacker a way to lock them out. Offer a review of enrolled credentials afterwards instead.

### 5. The operator, for self-hosted deployments

Whoever runs the server has database access, so they can always restore an account. That is a real escape hatch and should be written down rather than left implicit: **the operator of a deployment can take over any account on it.** Give them a command that does it deliberately and writes an audit row, rather than leaving them to hand-edit rows at 1am.

### Not in the ladder

| Approach | Why not |
|---|---|
| Security questions | The answers are public records or forgotten; they weaken every account that has them |
| SMS as the only path | SIM swap is routine and cheap |
| Unbounded support-mediated recovery | It becomes the attack, and it does not scale to a self-hosted deployment with no support desk |
| Identity documents, liveness checks | A data liability and a fraud surface of its own; out of scope for a library |
| No recovery at all | Defensible for a wallet. Cruel for a family photo album. |

## Rules for every path

- **Enumeration-resistant.** Starting recovery for an address that has no account looks exactly like starting one that does.
- **Bounded.** Attempts are counted per account across codes and messages, not per message, or requesting a fresh code resets the guess budget. Per-address and per-IP limits apply on top.
- **Loud.** A recovery attempt notifies every verified identifier and every live session. A completed one does the same and writes an audit row. A takeover people can see is a takeover they can report.
- **Held when the ground is fresh.** A hold — 24 to 72 hours, announced to every identifier — applies when recovery arrives through an identifier that was added or changed recently, which is the shape of an account takeover. It does **not** apply on the ordinary path of a long-verified address, because delaying that punishes the person you are trying to help.
- **Narrow.** Recovery adds a credential. It never changes identifiers, profile or anything else in the same step.
- **The last credential is protected.** An account cannot delete its only passkey unless recovery codes exist, because with no password to fall back on that is not an inconvenience, it is a permanently unreachable account.

## What breaks without each rule

| Skipped | Result |
|---|---|
| A second credential prompted at enrollment | Every account is one lost device from the recovery path |
| Attempts bounded per account | A six-digit code is guessable inside its own lifetime |
| Notifications | A takeover is silent, and nobody reports it |
| Hold on recently changed identifiers | The standard takeover works: change the address, recover through it |
| Recovery ungated by registration | A closed deployment locks out its own users forever |
| Not auto-revoking on recovery | An attacker who reaches recovery can also lock the owner out |
| The last-credential guard | Accounts delete themselves by accident |

## What a deployment chooses

Portico ships the ceremony, the codes, the limits, the notifications and the audit trail. A deployment decides the rest:

- whether recovery codes are on (they should be)
- which identifier kinds count, and how they are verified
- the hold policy and its window
- who may run the operator command, and how that is logged
- the wording of every message, since these are the most-read emails the system sends
