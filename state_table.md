| event                                     | instance state   | suspect exists   | action                                                                                   |
| ----------------------------------------- | ---------------- | ---------------- | ---------------------------------------------------------------------------------------- |
| convergence loop with cell missing        | running          | no               | move existing lrp to suspect and create a new one                                        |
| new instance reaches running state        | running          | yes              | remove suspect                                                                           |
| convergence loop with cell still missing  | running          | yes              | remove suspect (this shouldn't usually happen, since it's covered by the line above)     |
| convergence loop with cell still missing  | not running      | yes              | nothing to do                                                                            |
| convergence loop and cell is back         | running          | yes              | remove suspect                                                                           |
| convergence loop and cell is back         | not running      | yes              | remove instance and move suspect back to instance                                        |
