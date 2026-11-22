# Per-cell verdict shapes stay resolved

The archived-report repair shapes: the verdict word sits OUTSIDE the strike
in its own cell words, every cell carries a strike, and a routed row resolves
without any strike.

| #    | Task                                                          | Source    |
| ---- | ------------------------------------------------------------- | --------- |
| ~~6~~ | ~~task six~~ done — repo commit 891d200                        | ~~old.md~~ |
| ~~7~~ | ~~task seven~~ **Superseded — docs/status/newer.md.**          | ~~old.md~~ |
| ~~8~~ | ~~task eight~~ **Won't implement — outlived its report.**      | ~~old.md~~ |
| 9    | routed without a strike **→ TODO_LIST**                        | x.md      |
