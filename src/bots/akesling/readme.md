# How to use this

1) copy my main.go and strategy.go
2) remove all my strategies, and StategyStateRegistry values (including in AllStrategies and DetermineStrategies)
3) Implement your strategy and put its necessary state in StrategyStateRegistry


AllStrategies is used to know what to run the setup for

DetermineStrategies runs every turn to know which strategies to care about _this turn_

SetUp() is run once at the beginning of the game, IsApplicable() is used to figure out if a strategy should even be run, Run() actually runs the strategy

If no strategies apply... or all strategies return a nil move... it just passes