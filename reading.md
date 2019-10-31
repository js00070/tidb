# 代码阅读记录

## executor

executor.go里SelectionExec的Open函数, 最后一行`e.inputRow = e.inputIter.End()`为啥设成End补设成First? StreamAggExec的Open函数里也有同样的疑问

aggregate.go里StreamAggExec的consumeOneGroup函数, 执行了groupChecker.meetNewGroup之后, 在当前的row满足group条件之后, 为何是执行了consumeGroupRows?