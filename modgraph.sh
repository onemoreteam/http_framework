#!/bin/sh

go mod graph | modgraphviz | dot -Tpng -o modgraph.png
