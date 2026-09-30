# SDD ledger — plan: C:\Users\lavantien\.claude\plans\harmonize-all-the-documentations-silly-sutherland.md

Spec: the plan file above. No separate spec exists; rulings without spec fallback are provisional.

Setup rulings:
- Ruling: work on main directly — repo practice commits docs work straight to main (see git log), EnterWorktree only on explicit request — cost if wrong: revert via git.
- Ruling: plan has 6 commit units, treated as tasks 1-6 in the Commit split section.

Pre-flight: tasks 1, 5, 6 all touch README.md and DOCUMENTATION_ENFORCEMENT.md/scripts/README.md. Order fixed as plan's commit order; factual fixes (task 5) land before the voice pass (task 6) so the voice pass rewrites final text once.
