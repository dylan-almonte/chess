## 1. Eval module skeleton

- [x] 1.1 Add `src/eval.rs` with material constants (P=100, N=320, B=330, R=500, Q=900, K=0) and wire `mod eval` / `pub use evaluate` from `lib.rs` — verify with `cargo check`
- [ ] 1.2 Add failing named tests for all evaluation delta-spec scenarios (startpos `0`, side-to-move invariance, queen imbalances, e4>e2 pawn, mirrored kings `0`) — verify tests compile and fail before implementation (`cargo test evaluate_ -- --nocapture` shows failures)

## 2. Material and PST

- [ ] 2.1 Implement `evaluate(&Position) -> i32` material-only (White minus Black) — verify queen-imbalance scenarios pass (`> 800` / `< -800`)
- [ ] 2.2 Add mirrored PST tables (`sq ^ 56` for Black) and include them in the score — verify startpos `0`, side-to-move invariance, mirrored kings `0`, and e4 pawn score > e2 pawn score

## 3. Gate

- [ ] 3.1 Ensure every delta-spec scenario has a named `#[test]` and `cargo test` passes
- [ ] 3.2 Run `openspec validate evaluation --strict` and fix any issues
