# dice notation evaluator

This repository aims to evaluate TTRPG dice notation (e.g in D&D) used for rolling dice and evaluate them to a result.

## examples

|    | expression | feature                        | description                                      |
|:--:|:-----------|:-------------------------------|:-------------------------------------------------|
| ✅ | `d20`      | single dice                    | 1 roll of a 20 sided die                         |
| ✅ | `2d20`     | multiple rolls of same die     | 2 rolls of a 20 sided die                        |
| ✅ | `1d10+5`   | constant modifier              | 1 roll of a 20 sided die plus a coefficient of 5 |
| ✅ | `2d20kh1`  | keep highest *n* (advantage)   | 2 rolls of a 20 sided die, keep highest only     |
| ✅ | `2d20kl1`  | keep lowest *n* (disadvantage) | 2 rolls of a 20 sided die, keep lowest only      |
|    | `1d6!`     | exploding dice                 | 1 roll of a 6 sided die, re-roll if max value    |
|    | `4d6dl1`   | drop lowest *n*                | 4 rolls of a 6 sided die, discard lowest roll    |
