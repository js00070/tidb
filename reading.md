# 代码阅读记录

executor.go里SelectionExec的Open函数, 最后一行`e.inputRow = e.inputIter.End()`为啥设成End补设成First?