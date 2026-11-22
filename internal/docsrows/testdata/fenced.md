# Fenced examples are not corpus content

A real, fully resolved table:

| #    | Task                     | Source    |
| ---- | ------------------------ | --------- |
| ~~1~~ | ~~done task~~ done — landed | ~~a.md~~ |

An unresolved table shown as documentation inside a fenced block must be
ignored in every mode:

```markdown
| # | Task                | Source |
| - | ------------------- | ------ |
| 2 | still open example  | b.md   |
```

A fenced ledger marker is an example, not an exemption: the next real table
below stays judged. The partial row proves no exemption leaked:

```markdown
<!-- check-rows:ledger -->
| # | Task  | Source |
| - | ----- | ------ |
| 3 | open  | c.md   |
```

| #    | Task          | Source |
| ---- | ------------- | ----- |
| ~~4~~ | half struck  | d.md  |
