# Decision register (step 2)

`docs/decision-register.md` is the project's decision log. This project uses its own convention, not separate ADR files, except for larger choices, which also get an ADR in `docs/adr/`.

## Asking the owner

- Ask only after the research is done, and only for choices that are really the owner's to make. Editorial show values are studio settings, and conventional defaults are your call: state which you chose.
- Each question stands on its own. Give 2–4 options, put the recommended one first, and say in a line what each costs and where its evidence is (the research file).
- If the owner asks to clarify, ask what they want clarified. Don't re-ask the same question.

## Writing the row

- Add the row to the section it belongs in, with the next free ID (for example TS-16 under "E. Tech stack").
- Columns are `| ID | Decision | Options | Recommendation | Status |`. Under Decision, put the choice in bold plus what was rejected and why. Under Recommendation, cite the evidence (the research file and the key sources).
- Status is `**SETTLED YYYY-MM-DD**`, with a qualifier if adoption depends on a test, for example "(evaluate; adoption follows the load test)".
- A settled row is never edited to mean something else. A new row supersedes it, and the old row and its sources are kept (Nygard, "Documenting Architecture Decisions"; adr.github.io).
- Update everything the decision touches in the same branch: SPEC sections, `stack/` READMEs, older research items marked superseded, and PLAN.md tasks.
