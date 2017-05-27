#1/bin/bash
awk '{print ($1 + 1) " " $2}' compileVersion > compileVersion.tmp && mv compileVersion.tmp compileVersion && cat compileVersion
